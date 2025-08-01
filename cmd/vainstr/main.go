// Package main is a simple nostr key miner that uses the fast bitcoin secp256k1
// C library to derive npubs with specified prefix/infix/suffix strings present.
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"orly.dev/pkg/crypto/ec/bech32"
	"orly.dev/pkg/crypto/ec/secp256k1"
	"orly.dev/pkg/crypto/p256k"
	"orly.dev/pkg/encoders/bech32encoding"
	"orly.dev/pkg/utils/atomic"
	"orly.dev/pkg/utils/chk"
	"orly.dev/pkg/utils/interrupt"
	"orly.dev/pkg/utils/log"
	"orly.dev/pkg/utils/lol"
	"orly.dev/pkg/utils/qu"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
)

var prefix = append(bech32encoding.PubHRP, '1')

const (
	PositionBeginning = iota
	PositionContains
	PositionEnding
)

type Result struct {
	sec  []byte
	npub []byte
	pub  []byte
}

var args struct {
	String   string `arg:"positional" help:"the string you want to appear in the npub"`
	Position string `arg:"positional" default:"end" help:"[begin|contain|end] default: end"`
	Threads  int    `help:"number of threads to mine with - defaults to using all CPU threads available"`
}

func main() {
	lol.SetLogLevel("info")
	arg.MustParse(&args)
	if args.String == "" {
		_, _ = fmt.Fprintln(
			os.Stderr,
			`Usage: vainstr [--threads THREADS] [STRING [POSITION]]

Positional arguments:
  STRING                 the string you want to appear in the npub
  POSITION               [begin|contain|end] default: end

Options:
  --threads THREADS      number of threads to mine with - defaults to using all CPU threads available
  --help, -h             display this help and exit`,
		)
		os.Exit(0)
	}
	var where int
	canonical := strings.ToLower(args.Position)
	switch {
	case strings.HasPrefix(canonical, "begin"):
		where = PositionBeginning
	case strings.Contains(canonical, "contain"):
		where = PositionContains
	case strings.HasSuffix(canonical, "end"):
		where = PositionEnding
	}
	if args.Threads == 0 {
		args.Threads = runtime.NumCPU()
	}
	if err := Vanity(args.String, where, args.Threads); chk.T(err) {
		log.F.F("error: %s", err)
	}
}

func Vanity(str string, where int, threads int) (err error) {

	// check the string has valid bech32 ciphers
	for i := range str {
		wrong := true
		for j := range bech32.Charset {
			if str[i] == bech32.Charset[j] {
				wrong = false
				break
			}
		}
		if wrong {
			return fmt.Errorf(
				"found invalid character '%c' only ones from '%s' allowed\n",
				str[i], bech32.Charset,
			)
		}
	}
	started := time.Now()
	quit, shutdown := qu.T(), qu.T()
	resC := make(chan Result)
	interrupt.AddHandler(
		func() {
			// this will stop work if CTRL-C or Interrupt signal from OS.
			shutdown.Q()
		},
	)
	var wg sync.WaitGroup
	counter := atomic.NewInt64(0)
	for i := 0; i < threads; i++ {
		log.D.F("starting up worker %d", i)
		go mine(str, where, quit, resC, &wg, counter)
	}
	tick := time.NewTicker(time.Second * 5)
	var res Result
out:
	for {
		select {
		case <-tick.C:
			workingFor := time.Now().Sub(started)
			wm := workingFor % time.Second
			workingFor -= wm
			fmt.Printf(
				" working for %v, attempts %d",
				workingFor, counter.Load(),
			)
		case r := <-resC:
			// one of the workers found the solution
			res = r
			// tell the others to stop
			quit.Q()
			break out
		case <-shutdown.Wait():
			quit.Q()
			log.I.Ln("\rinterrupt signal received")
			os.Exit(0)
		}
	}

	// wait for all workers to stop
	wg.Wait()

	fmt.Printf(
		"\r# generated in %d attempts using %d threads, taking %v                                                 ",
		counter.Load(), args.Threads, time.Now().Sub(started),
	)
	fmt.Printf(
		"\nHSEC = %s\nHPUB = %s\n",
		hex.EncodeToString(res.sec),
		hex.EncodeToString(res.pub),
	)
	nsec, _ := bech32encoding.BinToNsec(res.sec)
	fmt.Printf("NSEC = %s\nNPUB = %s\n", nsec, res.npub)
	return
}

func mine(
	str string, where int, quit qu.C, resC chan Result, wg *sync.WaitGroup,
	counter *atomic.Int64,
) {

	wg.Add(1)
	var r Result
	var e error
	found := false
out:
	for {
		select {
		case <-quit:
			wg.Done()
			if found {
				// send back the result
				log.D.Ln("sending back result\n")
				resC <- r
				log.D.Ln("sent\n")
			} else {
				log.D.Ln("other thread found it\n")
			}
			break out
		default:
		}
		counter.Inc()
		// r.sec, r.pub, e = GenKeyPair()
		r.sec, r.pub, e = Gen()
		if e != nil {
			log.E.Ln("error generating key: '%v' worker stopping", e)
			break out
		}
		// r.npub, e = bech32encoding.PublicKeyToNpub(r.pub)
		if r.npub, e = bech32encoding.BinToNpub(r.pub); e != nil {
			log.E.Ln("fatal error generating npub: %s\n", e)
			break out
		}
		fmt.Printf("\rgenerating key: %s", r.npub)
		// log.I.F("%s", r.npub)
		switch where {
		case PositionBeginning:
			if bytes.HasPrefix(r.npub, append(prefix, []byte(str)...)) {
				found = true
				quit.Q()
			}
		case PositionEnding:
			if bytes.HasSuffix(r.npub, []byte(str)) {
				found = true
				quit.Q()
			}
		case PositionContains:
			if bytes.Contains(r.npub, []byte(str)) {
				found = true
				quit.Q()
			}
		}
	}
}

func Gen() (skb, pkb []byte, err error) {
	skb, pkb, _, _, err = p256k.Generate()
	return
}

// GenKeyPair creates a fresh new key pair using the entropy source used by
// crypto/rand (ie, /dev/random on posix systems).
func GenKeyPair() (
	sec *secp256k1.SecretKey,
	pub *secp256k1.PublicKey, err error,
) {

	sec, err = secp256k1.GenerateSecretKey()
	if err != nil {
		err = fmt.Errorf("error generating key: %s", err)
		return
	}
	pub = sec.PubKey()
	return
}

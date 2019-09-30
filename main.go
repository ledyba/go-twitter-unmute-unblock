package main

import (
	"flag"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"context"
	"github.com/ChimeraCoder/anaconda"
	"github.com/fatih/color"
	"github.com/ledyba/go-twitter-unmute-unblock/conf"
)

//go:generate bash geninfo.sh


func unmuteAll(wg *sync.WaitGroup, ctx context.Context, api *anaconda.TwitterApi) {
	defer wg.Done()
	var err error
	var cur anaconda.UserCursor

	cur, err = api.GetMutedUsersList(nil)
	if err != nil {
		log.Fatal(err)
	}

	cnt := 0
	for {
		for _, usr := range cur.Users {
			_, err = api.UnmuteUser(usr.ScreenName, nil)
			if err != nil {
				log.Errorf("Could not unmute [%s(%s@%d)]: %v", usr.Name, usr.ScreenName, usr.Id, err)
				continue
			}
			log.Infof("Unmute[%d]: %s(%s@%d)", cnt, usr.Name, usr.ScreenName, usr.Id)
			cnt++
		}
		param := url.Values{}
		param.Set("cursor", cur.Next_cursor_str)
		cur, err = api.GetMutedUsersList(param)
		if err != nil {
			log.Fatal(err)
		}
		if len(cur.Users) == 0{
			break
		}
	}
	log.Printf("Unmute done for %d users.", cnt)
}

func unblockAll(wg *sync.WaitGroup, ctx context.Context, api *anaconda.TwitterApi) {
	defer wg.Done()
	var err error
	var cur anaconda.UserCursor

	cur, err = api.GetBlocksList(nil)
	if err != nil {
		log.Fatal(err)
	}

	cnt := 0
	for {
		for _, usr := range cur.Users {
			_, err = api.UnblockUser(usr.ScreenName, nil)
			if err != nil {
				log.Errorf("Could not unblock [%s(%s@%d)]: %v", usr.Name, usr.ScreenName, usr.Id, err)
				continue
			}
			log.Infof("Unblock[%d]: %s(%s@%d)", cnt, usr.Name, usr.ScreenName, usr.Id)
			cnt++
		}
		param := url.Values{}
		param.Set("cursor", cur.Next_cursor_str)
		cur, err = api.GetBlocksList(param)
		if err != nil {
			log.Fatal(err)
		}
		if len(cur.Users) == 0 {
			break
		}
	}
	log.Printf("Unblock done for %d users.", cnt)
}

func mainLoop(sig <-chan os.Signal) os.Signal {
	anaconda.SetConsumerKey(conf.ConsumerKey)
	anaconda.SetConsumerSecret(conf.ConsumerSecret)
	api := anaconda.NewTwitterApi(conf.OAuthToken, conf.OAuthSecret)
	api.EnableThrottling(500*time.Millisecond, 5)
	//api.SetDelay(3 * time.Second)
	defer api.Close()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	waitChan := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go unmuteAll(wg, ctx, api)
	go unblockAll(wg, ctx, api)
	go func() {
		wg.Wait()
		close(waitChan)
	}()
	select {
	case <-waitChan:
		cancel()
		return nil
	case s := <-sig:
		cancel()
		return s
	}
}

func main() {
	//var err error

	log.Infof("Build at: %s", color.MagentaString("%s", buildAt()))
	log.Infof("Git Revision: \n%s", color.MagentaString("%s", gitRev()))
	flag.Parse()
	log.Info("----------------------------------------")
	log.Info("Initializing...")
	log.Info("----------------------------------------")

	log.Info(color.GreenString("                                    [OK]"))

	log.Info("----------------------------------------")
	log.Info("Initialized.")
	log.Info("----------------------------------------")

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := mainLoop(sig)
	log.Fatalf("Signal (%v) received, stopping\n", s)
}

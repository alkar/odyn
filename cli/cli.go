package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/alkar/odyn"
	"github.com/hashicorp/logutils"
	"github.com/jawher/mow.cli"
)

var (
	appVersion = "master"

	psCombinedTwo, _   = odyn.NewProviderSet(odyn.ProviderSetParallel, odyn.IpifyProvider, odyn.OpenDNSProvider)
	psCombinedThree, _ = odyn.NewProviderSet(odyn.ProviderSetSerial, psCombinedTwo, odyn.IPInfoProvider)

	publicipProviders = map[string]interface{}{
		"ipify":    odyn.IpifyProvider,
		"ipinfo":   odyn.IPInfoProvider,
		"opendns":  odyn.OpenDNSProvider,
		"combined": psCombinedThree,
	}

	dnsProviders = map[string]interface{}{
		"route53": nil,
	}
)

func init() {
	var err error
	dnsProviders["route53"], err = odyn.NewRoute53Zone()
	if err != nil {
		log.Printf("[ERROR] error initialising route53: %+v", err)
		os.Exit(1)
	}
}

func getPublicIPProvider(name string) odyn.IPProvider {
	return validateProvider(name, publicipProviders).(odyn.IPProvider)
}

func getDNSZoneProvider(name string) odyn.DNSZone {
	return validateProvider(name, dnsProviders).(odyn.DNSZone)
}

func validateProvider(name string, providers map[string]interface{}) interface{} {
	for k, v := range providers {
		if k == name {
			return v
		}
	}

	available := make([]string, len(providers))
	i := 0
	for k := range providers {
		available[i] = k
		i++
	}
	fmt.Printf("[ERROR] invalid value '%s': provider must be one of: %s", name, strings.Join(available, ", "))
	os.Exit(1)

	return nil
}

func initLog(debug bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stderr,
	}
	if debug {
		filter.MinLevel = logutils.LogLevel("DEBUG")
	}
	log.SetOutput(filter)
}

func main() {
	var (
		app              = cli.App("odyn", "Odyn is a modern, extensible dynamic DNS updater")
		debugLog         = app.BoolOpt("d debug", false, "enables debug log output")
		publicIPProvider = app.StringOpt("p public-ip-provider", "combined", "public IP provider to use")
		dnsZoneProvider  = app.StringOpt("d dns-zone-provider", "route53", "DNS provider to use")
		zoneName         = app.StringArg("ZONE", "", "DNS zone")
		recordName       = app.StringArg("RECORD", "", "DNS record to update")
	)

	app.Action = func() {
		initLog(*debugLog)

		publicIP := getPublicIPProvider(*publicIPProvider)
		dnsZone := getDNSZoneProvider(*dnsZoneProvider)
		u := newUpdater(*recordName, *zoneName, publicIP, dnsZone)

		sigChannel := make(chan os.Signal, 1)
		signal.Notify(sigChannel, os.Interrupt)
		go func() {
			<-sigChannel
			log.Println("[INFO] interrupt singal: shutting down ...")
			u.stop()
		}()

		u.start()
	}

	app.Version("v version", appVersion)

	app.Run(os.Args)
}

type updater struct {
	*odyn.DNSClient
	odyn.IPProvider
	odyn.DNSZone
	zoneName   string
	recordName string
	stopChan   chan struct{}
}

func newUpdater(recordName, zoneName string, ipProvider odyn.IPProvider, dnsZone odyn.DNSZone) *updater {
	return &updater{
		odyn.NewDNSClient(),
		ipProvider,
		dnsZone,
		zoneName,
		recordName,
		make(chan struct{}),
	}
}

func (u *updater) start() {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	u.sync()
	for {
		select {
		case <-tick.C:
			u.sync()
		case <-u.stopChan:
			return
		}
	}
}

func (u *updater) stop() {
	close(u.stopChan)
}

func (u *updater) sync() {
	zoneNameservers, err := u.Nameservers(u.zoneName)
	if err != nil {
		log.Printf("[ERROR] could not get dns zone's nameservers: %+v", err)
		return
	}
	for i := 0; i < len(zoneNameservers); i++ {
		zoneNameservers[i] += ":53"
	}

	ipRecord, err := u.ResolveA(u.recordName, zoneNameservers)
	if err != nil {
		log.Printf("[INFO] could not resolve current DNS record, ignoring error: %+v", err)
		return
	}
	if len(ipRecord) > 1 {
		log.Printf("[INFO] nameserver replied with multiple IP addresses, will use the first: %+v", ipRecord)
	}

	ipCurrent, err := u.Get()
	if err != nil {
		log.Printf("[ERROR] could not get public IP address: %+v", err)
		return
	}

	if ipCurrent.Equal(ipRecord[0]) {
		log.Printf("[DEBUG] current public IP address is already registered with the nameservers, will not update")
		return
	}

	log.Printf("[INFO] IP address has changed, updating ...")
	err = u.UpdateA(u.recordName, u.zoneName, ipCurrent)
	if err != nil {
		log.Printf("[ERROR] failed to update the DNS record, will try again in roughly a minute: %+v", err)
		return
	}
	log.Printf("[INFO] updated the DNS record to point to: %+v", ipCurrent)
}

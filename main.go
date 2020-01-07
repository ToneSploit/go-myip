package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

/*
	http://www.sorbs.net/using.shtml

	http.dnsbl.sorbs.net    127.0.0.2
	socks.dnsbl.sorbs.net    127.0.0.3
	misc.dnsbl.sorbs.net    127.0.0.4
	smtp.dnsbl.sorbs.net    127.0.0.5
	new.spam.dnsbl.sorbs.net    127.0.0.6
	recent.spam.dnsbl.sorbs.net    127.0.0.6
	old.spam.dnsbl.sorbs.net    127.0.0.6
	spam.dnsbl.sorbs.net    127.0.0.6
	escalations.dnsbl.sorbs.net    127.0.0.6
	web.dnsbl.sorbs.net    127.0.0.7
	block.dnsbl.sorbs.net    127.0.0.8
	zombie.dnsbl.sorbs.net    127.0.0.9
	dul.dnsbl.sorbs.net    127.0.0.10
	badconf.rhsbl.sorbs.net    127.0.0.11
	nomail.rhsbl.sorbs.net    127.0.0.12
	noserver.dnsbl.sorbs.net    127.0.0.14
	virus.dnsbl.sorbs.net    127.0.0.15
*/

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(colorable.NewColorableStdout())

	// logrus.Info("succeeded")
	// logrus.Warn("not correct")
	// logrus.Error("something error")
	// logrus.Fatal("panic")

	file, err := os.Open("ips.txt")
	if err != nil {
		logrus.Fatal(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		rev, err := reverseIPv4(scanner.Text())
		if err != nil {
			logrus.Warn("IP not correct")
		}

		response, err := GetA(rev+".dnsbl.sorbs.net", "8.8.8.8")
		if err != nil {
			logrus.Warn("IP not correct:" + err.Error())
		}
		logrus.WithFields(logrus.Fields{
			"IP":       scanner.Text(),
			"Record":   rev + ".dnsbl.sorbs.net",
			"Response": response,
		}).Info("done")

		// logrus.Warn(len(response))
	}

	if err := scanner.Err(); err != nil {
		logrus.Fatal(err.Error())
	}

	// ipAddress := "127.0.0.2"

}

func GetA(hostname string, nameserver string) ([]string, error) {
	var record []string
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(hostname), dns.TypeA)
	c := new(dns.Client)
	m.MsgHdr.RecursionDesired = true
	in, _, err := c.Exchange(m, nameserver+":53")
	if err != nil {
		return nil, err
	}
	for _, rin := range in.Answer {
		if r, ok := rin.(*dns.A); ok {
			record = append(record, r.A.String())
		}
	}

	return record, nil
}

func reverseIPv4(ip string) (string, error) {
	PTR, err := dns.ReverseAddr(ip)
	if err != nil {
		return "", err
	}

	reversed := strings.TrimSuffix(PTR, ".in-addr.arpa.")

	return reversed, nil
}

/*
var Blacklists = []string{
	"aspews.ext.sorbs.net",
	"b.barracudacentral.org",
	"bl.deadbeef.com",
	"bl.emailbasura.org",
	"bl.spamcannibal.org",
	"bl.spamcop.net",
	"blackholes.five-ten-sg.com",
	"blacklist.woody.ch",
	"bogons.cymru.com",
	"cbl.abuseat.org",
	"cdl.anti-spam.org.cn",
	"combined.abuse.ch",
	"combined.rbl.msrbl.net",
	"db.wpbl.info",
	"dnsbl-1.uceprotect.net",
	"dnsbl-2.uceprotect.net",
	"dnsbl-3.uceprotect.net",
	"dnsbl.cyberlogic.net",
	"dnsbl.dronebl.org",
	"dnsbl.inps.de",
	"dnsbl.njabl.org",
	"dnsbl.sorbs.net",
	"drone.abuse.ch",
	"duinv.aupads.org",
	"dul.dnsbl.sorbs.net",
	"dul.ru",
	"dyna.spamrats.com",
	"dynip.rothen.com",
	"http.dnsbl.sorbs.net",
	"images.rbl.msrbl.net",
	"ips.backscatterer.org",
	"ix.dnsbl.manitu.net",
	"korea.services.net",
	"misc.dnsbl.sorbs.net",
	"noptr.spamrats.com",
	"ohps.dnsbl.net.au",
	"omrs.dnsbl.net.au",
	"orvedb.aupads.org",
	"osps.dnsbl.net.au",
	"osrs.dnsbl.net.au",
	"owfs.dnsbl.net.au",
	"owps.dnsbl.net.au",
	"pbl.spamhaus.org",
	"phishing.rbl.msrbl.net",
	"probes.dnsbl.net.au",
	"proxy.bl.gweep.ca",
	"proxy.block.transip.nl",
	"psbl.surriel.com",
	"rdts.dnsbl.net.au",
	"relays.bl.gweep.ca",
	"relays.bl.kundenserver.de",
	"relays.nether.net",
	"residential.block.transip.nl",
	"ricn.dnsbl.net.au",
	"rmst.dnsbl.net.au",
	"sbl.spamhaus.org",
	"short.rbl.jp",
	"smtp.dnsbl.sorbs.net",
	"socks.dnsbl.sorbs.net",
	"spam.abuse.ch",
	"spam.dnsbl.sorbs.net",
	"spam.rbl.msrbl.net",
	"spam.spamrats.com",
	"spamlist.or.kr",
	"spamrbl.imp.ch",
	"t3direct.dnsbl.net.au",
	"tor.dnsbl.sectoor.de",
	"torserver.tor.dnsbl.sectoor.de",
	"ubl.lashback.com",
	"ubl.unsubscore.com",
	"virbl.bit.nl",
	"virus.rbl.jp",
	"virus.rbl.msrbl.net",
	"web.dnsbl.sorbs.net",
	"wormrbl.imp.ch",
	"xbl.spamhaus.org",
	"zen.spamhaus.org",
	"zombie.dnsbl.sorbs.net"}

*/

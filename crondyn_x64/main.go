/*
	crondyn, a cron-compatible dynamic DNS update utility.
	© 2014 Ken Piper <kealper@gmail.com>
	
    This file is part of crondyn.

    crondyn is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    crondyn is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with crondyn.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	"io/ioutil"
	"runtime"
)

type Config struct {
	Username string
	Password string
	Hostname string
	DNSServer string
}

func shutDown() {
	if panicReason := recover(); panicReason != nil {
		fmt.Println("Program has encountered an unrecoverable error and has crashed.")
		fmt.Println("Some information describing this crash: "+panicReason.(error).Error())
		fmt.Println("Crash reporting is enabled, collecting stack trace and submitting crash report...")
		stack := make([]byte, 8192)
		l := runtime.Stack(stack, true)
		resp, err := http.PostForm("http://stacktrace.kealper.com/api.a3w", url.Values{"agent": {"crondns/1.0"}, "crash": {panicReason.(error).Error()}, "stack": {string(stack[:l])}})
		if err != nil {
			fmt.Println("Failed to submit crash report, closing.")
		} else {
			resp.Body.Close()
			fmt.Println("Crash report submitted, closing.")
		}
	}
}

func main() {
	defer shutDown()
	configFile, _ := ioutil.ReadFile("config.json")
	config := new(Config)
	err := json.Unmarshal(configFile, config)
	if err != nil {
		fmt.Println("Failed to load configuration! Closing...")
		return
	}
	lastIP, err := ioutil.ReadFile("lastip")
	if err != nil {
		lastIP = []byte("0.0.0.0")
	}
	resp, err := http.Get("http://"+config.DNSServer+"/checkip.php")
	if err != nil {
		resp, err = http.Get("http://kealper.com/ip")
		if err != nil {
			fmt.Println("Failed to get current IP address! Closing...")
			return
		}
	}
	currentIP, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if string(currentIP) != string(lastIP) {
		fmt.Println("Found updated IP address:", string(currentIP))
		if err := ioutil.WriteFile("lastip", currentIP, 0644); err != nil {
			fmt.Println("Failed to save the current IP to the disk! Permissions issue?")
		}
		_, err := http.Get("http://"+config.Username+":"+config.Password+"@"+config.DNSServer+"/dynamicupdate.php?ip="+string(currentIP)+"&hostname="+config.Hostname)
		if err != nil {
			fmt.Println("Failed to update hostname:", config.Hostname)
			return
		} else {
			fmt.Println("Hostname", config.Hostname, "updated successfully.")
		}
	} else {
		fmt.Println("No IP address change since last check.")
	}
}

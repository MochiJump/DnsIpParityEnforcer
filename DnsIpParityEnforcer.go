package main

import (
    "encoding/json"
    "fmt"
    "github.com/go-co-op/gocron"
    "io/ioutil"
    "net/http"
    "time"
    "net"
)

var prevIp = "0.0.0.0"

func main() {
    fmt.Println("Creating DNS IP Parity Enforcer Chron Job")
    scheduler := gocron.NewScheduler(time.UTC)
    scheduler.Every(10).Minutes().Do(EnforceIpParity)
    scheduler.StartBlocking()
}

type ApiResp struct{
    Ip string
    Country string
}

func EnforceIpParity() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("chron job failed in an unexpected way", r)
        }
    }()
    if prevIp == "0.0.0.0"{
        fmt.Println("Previous DNS Record Unknown, attempting to find IP associated with DNS record")
        findCurrentDnsIp()
    }
    body, err := queryForIpAddress()
    if err != nil{
        return
    }
    var resp ApiResp
    json.Unmarshal(body, &resp)
    if prevIp == resp.Ip {
        fmt.Println("Ip address is the same. No change to DNS required. IP is ", resp.Ip)
    } else{
        changeWebsiteIpAddress(resp)
    }

}

func findCurrentDnsIp(){
    ips, err := net.LookupIP("mochijump.com")
    if err != nil {
        fmt.Println("Error %s", err)
        return
    }
    for _, ip := range ips {
        // currently restricted to ipv4
        if ipv4 := ip.To4(); ipv4 != nil {
            fmt.Println("Dns Ip Address is Currently Set to: ", ipv4)
            prevIp = fmt.Sprintf("%s", ipv4)
        }
    }
}


func queryForIpAddress() ([]byte, error) {
    resp, err := http.Get("https://api.myip.com")
    if err != nil {
        fmt.Println("Error %s", err)
        return nil, err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error %s", err)
        return nil, err
    }
    return body, nil
}

func changeWebsiteIpAddress(resp ApiResp){
    fmt.Println("Changing ip address in DNS record to", resp.Ip)
    host := "mochijump.com"
    url := "https://user:pass@domains.google.com/nic/update?hostname="+host+"&myip="+resp.Ip
    res, err := http.Get(url)
    if err != nil {
        fmt.Println("Error %s", err)
        return
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println("Error %s", err)
        return
    }
    //fmt.Println("HTTP status:", res.StatusCode)
    fmt.Println(string(body))
    if string(body) == "badauth" {
        fmt.Println("failed to update dns")
        return
    }

    prevIp = resp.Ip
}

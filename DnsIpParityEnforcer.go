package main

import (
    "encoding/json"
    "fmt"
    "github.com/go-co-op/gocron"
    "io/ioutil"
    "net/http"
    "time"
)

var prevIp = "0.0.0.0"

func main() {

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
    var resp ApiResp
    body, err := queryForIpAddress()
    if err != nil{
        return
    }
    json.Unmarshal(body, &resp)
    if prevIp == resp.Ip {
        fmt.Println("Ip address is the same. No change to DNS required")
    } else{
        changeWebsiteIpAddress(resp)
    }

}

func queryForIpAddress() ([]byte, error) {
    resp, err := http.Get("https://api.myip.com")
    if err != nil {
        fmt.Printf("Error %s", err)
        return nil, err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error %s", err)
        return nil, err
    }
    return body, nil
}

func changeWebsiteIpAddress(resp ApiResp){
    fmt.Println("Changing ip address in DNS record to", resp.Ip)
    url := "https://username:password@domains.google.com/nic/update?hostname=subdomain.domain.com&myip="+resp.Ip
    res, err := http.Get(url)
    if err != nil {
        fmt.Printf("Error %s", err)
        return
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Printf("Error %s", err)
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

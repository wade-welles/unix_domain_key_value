package access

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ldkingvivi/unix_domain_key_value/unix"
)

//keyStore ...
type keyStore struct {
	Name      string
	Timestamp string
	Version   string
	Content   map[string]bool
}

//Data ...
type Data struct {
	map1           keyStore
	map2           keyStore
	currentmap     *keyStore
	updatemap      *keyStore
	mapswitch      chan string
	mapupdate      chan string
	initupdate     chan string
	UConnChan      chan *unix.UConn
	sourceendpoint string
	updateInterval int
	apitoken       string
	apiClient      *http.Client
	apiReq         *http.Request
}

//InitMap ...
func InitMap(UConnChan chan *unix.UConn, updateInterval int, sourceendpoint string, apitoken string) (*Data, error) {

	d := new(Data)
	d.map1.Name = "map1"
	d.map1.Content = make(map[string]bool)
	d.map2.Name = "map2"
	d.map2.Content = make(map[string]bool)

	d.currentmap = &d.map1
	d.updatemap = &d.map2

	d.mapswitch = make(chan string)
	d.mapupdate = make(chan string, 10)
	d.initupdate = make(chan string)
	d.UConnChan = UConnChan
	d.updateInterval = updateInterval
	d.sourceendpoint = sourceendpoint
	d.apitoken = "Bearer " + apitoken
	d.apiClient = &http.Client{}
	req, err := http.NewRequest("GET", d.sourceendpoint, nil)
	if err != nil {
		os.Exit(1)
	}
	d.apiReq = req
	d.apiReq.Header.Add("Authorization", d.apitoken)

	return d, nil
}

func (d *Data) update() {
	log.Printf("start to update map: %s\n", d.updatemap.Name)

	d.updatemap.Content = nil
	d.updatemap.Content = make(map[string]bool)

	resp, err := d.apiClient.Do(d.apiReq)
	if err != nil {
		//some error, send signal to updtae again, will wait for another cycle
		d.mapupdate <- "update"
		return
	}

	if resp != nil {
		//get resp
		errDecode := json.NewDecoder(resp.Body).Decode(&d.updatemap.Content)

		if errDecode != nil {
			d.mapupdate <- "update"
			return
		}
		resp.Body.Close()
		d.mapswitch <- "switch"

	} else {
		d.mapupdate <- "update"
	}
	return
}

//keepUpdate ...
func (d *Data) keepUpdate() {
	for {
		select {
		case <-d.mapupdate:
			time.Sleep(time.Duration(d.updateInterval) * time.Second)
			d.update()

		case <-d.initupdate:
			d.update()
		}
	}
}

//rwHandler, this intend to solve the rw issue without lock
func (d *Data) rwHandler() {

	var tempmap *keyStore

	for {
		select {
		case <-d.mapswitch:

			tempmap = d.currentmap
			d.currentmap = d.updatemap
			d.updatemap = tempmap
			d.mapupdate <- "update"

		case uConnQuery := <-d.UConnChan:

			lIPDest := strings.Split(uConnQuery.Query, "|")
			if len(lIPDest) < 2 {
				uConnQuery.Qchan <- "false\r\n"
			} else {
				ipSpecificDest := lIPDest[0] + "_" + lIPDest[1]

				_, okSpecific := d.currentmap.Content[ipSpecificDest]

				if okSpecific {
					uConnQuery.Qchan <- "true\r\n"
				} else {
					uConnQuery.Qchan <- "false\r\n"
				}
			}
		}
	}
}

//Start ...
func (d *Data) Start() {

	go d.rwHandler()
	go d.keepUpdate()

	d.initupdate <- "start"

}

package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	Registrations []Registration
	mutex *sync.RWMutex
}

//r *registry代表可以修改registry中的内容
func (r *registry) add(reg Registration) error {
	r.mutex.Lock()
	r.Registrations = append(r.Registrations, reg)
	r.mutex.Unlock()
	err := r.sendRequiredService(reg)
	r.notify(patch{
		Added: []patchEntry{
			patchEntry{
				Name: reg.ServiceName,
				URL: reg.ServiceURL,
			},
		},
	})
	return err
}

func(r registry) notify(fullPatch patch) { //fullPatch为本次更新的服务
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, reg := range r.Registrations {  //针对已经注册的服务进行循环
		go func(reg Registration) {
			for _, reqService := range reg.RequiredServices { //针对每个已注册服务所依赖的服务进行循环
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				sendUpdate := false
				for _, added := range fullPatch.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
}

func (r registry) sendRequiredService(reg Registration) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	for _, serviceReg := range r.Registrations {
		 for _, reqService := range reg.RequiredServices {
			  if serviceReg.ServiceName == reqService {  // 仅当所依赖的服务已经注册，才会向自身的updateURL发送一个依赖的patch，从而由client向provider写入信息，并由自身的main函数输出找到依赖服务的消息
				  p.Added = append(p.Added, patchEntry{
					  Name: serviceReg.ServiceName,
					  URL: serviceReg.ServiceURL,
				  })
			  }
		 }
	}
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil
}

func (r registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	return nil
}

func (r *registry) remove(url string) error {
	for i := range reg.Registrations {
		if reg.Registrations[i].ServiceURL == url {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: r.Registrations[i].ServiceName,
						URL: r.Registrations[i].ServiceURL,
					},
				},
			})
			r.mutex.Lock()
			r.Registrations = append(reg.Registrations[:i], reg.Registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("Service at URL %s not found", url)
}

var reg = registry{
	Registrations: make([]Registration, 0),
	mutex: new(sync.RWMutex),
}

type RegistryService struct {}

//s RegistryService代表不修改RegistryService中的内容
func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v with URL: %s\n", r.ServiceName, r.ServiceURL)
		err = reg.add(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL %s", url)
		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

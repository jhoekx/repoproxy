package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type proxy struct {
	mirror *url.URL
	rpmDir string
	tmpDir string
	lock   *sync.Mutex
}

func (p proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if strings.Contains(path, ".rpm") {
		p.mirrorFrom(p.rpmDir, w, req)
	} else if containsAny(path, "LiveOS", "isolinux", "images") {
		p.mirrorFrom(p.tmpDir, w, req)
	} else {
		p.proxy(w, req.URL)
	}
}

func containsAny(s string, substr ...string) bool {
	parts := strings.Split(s, "/")
	for _, part := range parts {
		for _, subs := range substr {
			if strings.Contains(part, subs) {
				return true
			}
		}
	}
	return false
}

func (p proxy) proxy(w http.ResponseWriter, url *url.URL) {
	log.Println("Proxying ", url)
	resp, err := http.Get(p.mirror.ResolveReference(url).String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println(err.Error())
	}
}

func (p proxy) mirrorFrom(dir string, w http.ResponseWriter, req *http.Request) {
	name := path.Join(dir, path.Clean(req.URL.Path))
	err := p.ensureFileExists(name, req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Serving", req.URL)
	http.ServeFile(w, req, name)
}

func (p proxy) ensureFileExists(name string, url *url.URL) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, err = os.Stat(name)
	if err == nil || !os.IsNotExist(err) {
		return
	}

	log.Println("Mirroring", url)
	resp, err := http.Get(p.mirror.ResolveReference(url).String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dirname := filepath.Dir(name)
	err = os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		return
	}
	f, err := os.Create(name)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return
}

type proxyConfig struct {
	mirrorFlag *string
	rpmDirFlag *string
}

func newProxyConfig(mirrorFlag *string, rpmDirFlag *string) (c proxyConfig) {
	c.mirrorFlag = mirrorFlag
	c.rpmDirFlag = rpmDirFlag
	return
}

func (c proxyConfig) build() (p proxy, err error) {
	mirror, err := c.getMirror()
	if err != nil {
		return
	}
	rpmDir, err := c.getRPMDir()
	if err != nil {
		return
	}
	tmpDir, err := ioutil.TempDir("", "mirror-tmp")
	if err != nil {
		return
	}
	p.mirror = mirror
	p.rpmDir = rpmDir
	p.tmpDir = tmpDir
	p.lock = &sync.Mutex{}
	return
}

func (c proxyConfig) getMirror() (url *url.URL, err error) {
	mirror := os.Getenv("CENTOS_MIRROR")
	if mirror == "" {
		mirror = *c.mirrorFlag
	}
	url, err = url.Parse(mirror)
	return
}

func (c proxyConfig) getRPMDir() (dir string, err error) {
	dir = os.Getenv("RPM_DIR")
	if dir != "" {
		return
	}

	dir = *c.rpmDirFlag
	if dir != "" {
		return
	}

	dir, err = ioutil.TempDir("", "mirror-rpms")
	return
}

func main() {
	var mirrorArg = flag.String("mirror", "http://centos.mirror.nucleus.be", "The CentOS mirror to proxy")
	var rpmDir = flag.String("rpmdir", "", "The directory to cache data in")
	flag.Parse()

	proxy, err := newProxyConfig(mirrorArg, rpmDir).build()
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", proxy)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

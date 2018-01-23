package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"gopkg.in/yaml.v2"
)

type Cache struct {
	resourceURL          *url.URL
	LastSuccessfulCommit string `yaml:"last_successful_commit"`
}

func NewCache(resourceURI string) (*Cache, error) {
	resourceURL, err := url.Parse(resourceURI)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s is not a correct url", resourceURI))
	}

	resource, err := openResource(resourceURL)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[WARNING] Cache file at %s not found, building all projects...", resourceURI)
			return nil, nil
		}

		return nil, err
	}
	defer resource.Close()

	resourceData, err := ioutil.ReadAll(resource)

	var cache Cache
	err = yaml.Unmarshal(resourceData, &cache)
	if err != nil {
		return nil, err
	}

	if cache.LastSuccessfulCommit == "" {
		return nil, nil
	}

	cache.resourceURL = resourceURL

	return &cache, nil
}

func (c *Cache) Commit() error {
	resource, err := openResource(c.resourceURL)
	if err != nil {
		return err
	}
	defer resource.Close()

	cacheData, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	_, err = resource.Write(cacheData)
	return err
}

func openResource(uri *url.URL) (io.ReadWriteCloser, error) {
	switch uri.Scheme {
	case "file":
		return os.Open(uri.Path)
	default:
		return nil, errors.New(fmt.Sprintf("Unsupported resource scheme %s", uri.Scheme))
	}
}

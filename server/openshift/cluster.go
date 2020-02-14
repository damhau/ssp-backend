package openshift

import (
	"errors"
	"log"
	"net/http"

	"github.com/SchweizerischeBundesbahnen/ssp-backend/server/config"
	"github.com/gin-gonic/gin"
)

type OpenshiftCluster struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Optgroup    string   `json:"optgroup"`
	Stages      []string `json:"stages"`
	Recommended bool     `json:"recommended"`
	// exclude token from json marshal
	Token      string      `json:"-"`
	URL        string      `json:"url"`
	GlusterApi *GlusterApi `json:"-"`
	NfsApi     *NfsApi     `json:"-"`
}

type GlusterApi struct {
	URL          string `json:"url"`
	Secret       string `json:"-"`
	IPs          string `json:"-"`
	StorageClass string `json:"-"`
}

type NfsApi struct {
	URL          string `json:"url"`
	Secret       string `json:"-"`
	Proxy        string `json:"-"`
	StorageClass string `json:"-"`
}

func clustersHandler(c *gin.Context) {
	clusters := getOpenshiftClusters(c.Query("stage"))
	// if we filter by stage, then the user is creating a new project
	if c.Query("stage") != "" {
		// This function directly modifies the clusters array
		// We ignore errors, because then we just do not recommend a cluster
		setRecommendedCluster(clusters)
	}
	c.JSON(http.StatusOK, clusters)
}

func getOpenshiftClusters(stage string) []OpenshiftCluster {
	clusters := []OpenshiftCluster{}
	config.Config().UnmarshalKey("openshift", &clusters)
	if stage != "" {
		tmp := []OpenshiftCluster{}
		for _, c := range clusters {
			if contains(c.Stages, stage) {
				tmp = append(tmp, c)
			}
		}
		return tmp
	}
	return clusters
}

func contains(list []string, search string) bool {
	for _, element := range list {
		if element == search {
			return true
		}
	}
	return false
}

func getOpenshiftCluster(clusterId string) (OpenshiftCluster, error) {
	if clusterId == "" {
		log.Printf("WARNING: clusterId missing!")
		return OpenshiftCluster{}, errors.New(genericAPIError)
	}
	clusters := getOpenshiftClusters("")
	for _, cluster := range clusters {
		if cluster.ID == clusterId {
			return cluster, nil
		}
	}
	log.Printf("WARNING: Cluster %v not found", clusterId)
	return OpenshiftCluster{}, errors.New(genericAPIError)
}

func getStorageClass(clusterId, technology string) (string, error) {

	cluster, err := getOpenshiftCluster(clusterId)
	if err != nil {
		return "", err
	}
	var storageclass string

	if technology == "nfs" {
		if cluster.NfsApi == nil {
			log.Printf("WARNING: NfsApi is not configured for cluster %v", clusterId)
			return "", nil
		}
		storageclass = cluster.NfsApi.StorageClass

	} else {
		if cluster.GlusterApi == nil {
			log.Printf("WARNING: GlusterApi is not configured for cluster %v", clusterId)
			return "", nil
		}

		storageclass = cluster.GlusterApi.StorageClass
	}
	return storageclass, nil
}

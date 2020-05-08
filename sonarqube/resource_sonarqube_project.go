package sonarqube

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	log "github.com/sirupsen/logrus"
)

// Returns the resource represented by this file.
func resourceSonarqubeProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceSonarqubeProjectCreate,
		Read:   resourceSonarqubeProjectRead,
		Delete: resourceSonarqubeProjectDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSonarqubeProjectImport,
		},

		// Define the fields of this schema.
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "public",
				ForceNew: true,
			},
		},
	}
}

func resourceSonarqubeProjectCreate(d *schema.ResourceData, m interface{}) error {
	url := m.(*ProviderConfiguration).url
	url.Path = "api/projects/create"
	url.ForceQuery = true
	url.RawQuery = fmt.Sprintf("name=%s&project=%s&visibility=%s",
		d.Get("name").(string),
		d.Get("project").(string),
		d.Get("visibility").(string),
	)

	req, err := http.NewRequest("POST", url.String(), http.NoBody)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectCreate")
		return err
	}
	resp, err := m.(*ProviderConfiguration).httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectCreate")
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		responseBody := getResponseBodyAsString(resp)
		return errors.New(responseBody)
	}

	projectResponse := CreateProjectResponse{}
	err = json.NewDecoder(resp.Body).Decode(&projectResponse)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectCreate")
	}

	d.SetId(projectResponse.Project.Key)
	return nil
}

func resourceSonarqubeProjectRead(d *schema.ResourceData, m interface{}) error {
	url := m.(*ProviderConfiguration).url
	url.Path = "api/projects/search"
	url.ForceQuery = true
	url.RawQuery = fmt.Sprintf("projects=%s",
		d.Id(),
	)

	req, err := http.NewRequest("GET", url.String(), http.NoBody)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectRead")
		return err
	}
	resp, err := m.(*ProviderConfiguration).httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectRead")
		return err
	}

	defer resp.Body.Close()
	log.WithField("status code", resp.StatusCode).Info("Response from server")
	if resp.StatusCode != http.StatusOK {
		responseBody := getResponseBodyAsString(resp)
		return errors.New(responseBody)
	}

	projectReadResponse := GetProject{}
	err = json.NewDecoder(resp.Body).Decode(&projectReadResponse)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectRead")
	}

	for _, value := range projectReadResponse.Components {
		if d.Id() == value.Key {
			d.SetId(value.Key)
			d.Set("name", value.Name)
			d.Set("key", value.Key)
			d.Set("visibility", value.Visibility)
		}
	}

	return nil
}

func resourceSonarqubeProjectDelete(d *schema.ResourceData, m interface{}) error {
	url := m.(*ProviderConfiguration).url
	url.Path = "api/projects/delete"
	url.ForceQuery = true
	url.RawQuery = fmt.Sprintf("projects=%s",
		d.Id(),
	)
	req, err := http.NewRequest("POST", url.String(), http.NoBody)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectDelete")
		return err
	}
	resp, err := m.(*ProviderConfiguration).httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("resourceSonarqubeProjectDelete")
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		responseBody := getResponseBodyAsString(resp)
		return errors.New(responseBody)
	}

	return nil
}

func resourceSonarqubeProjectImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceSonarqubeProjectRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

package github

import (
	"github.com/google/go-github/github"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGithubTeam() *schema.Resource {

	return &schema.Resource{
		Create: resourceGithubTeamCreate,
		Read:   resourceGithubTeamRead,
		Update: resourceGithubTeamUpdate,
		Delete: resourceGithubTeamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"privacy": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "secret",
				ValidateFunc: validateValueFunc([]string{"secret", "closed"}),
			},
			"etag": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubTeamCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).Client()
	n := d.Get("name").(string)
	desc := d.Get("description").(string)
	p := d.Get("privacy").(string)
	githubTeam, _, err := client.Organizations.CreateTeam(meta.(*Organization).name, &github.Team{
		Name:        &n,
		Description: &desc,
		Privacy:     &p,
	})
	if err != nil {
		return err
	}
	d.SetId(fromGithubID(githubTeam.ID))
	return resourceGithubTeamRead(d, meta)
}

func resourceGithubTeamRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).Client()

	team, err, etag := getGithubTeam(d, client)
	if err != nil {
		d.SetId("")
		return nil
	}
	d.Set("description", team.Description)
	d.Set("name", team.Name)
	d.Set("privacy", team.Privacy)
	d.Set("etag", etag)
	return nil
}

func resourceGithubTeamUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).Client()

	team, err, _ := getGithubTeam(d, client)

	if err != nil {
		d.SetId("")
		return nil
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	privacy := d.Get("privacy").(string)
	team.Description = &description
	team.Name = &name
	team.Privacy = &privacy

	team, _, err = client.Organizations.EditTeam(*team.ID, team)
	if err != nil {
		return err
	}
	d.SetId(fromGithubID(team.ID))
	return resourceGithubTeamRead(d, meta)
}

func resourceGithubTeamDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).Client()
	id := toGithubID(d.Id())
	_, err := client.Organizations.DeleteTeam(id)
	return err
}

func getGithubTeam(d *schema.ResourceData, client *Client) (*github.Team, error, string) {
	id := toGithubID(d.Id())
	client.Transport.etag = d.Get("etag").(string)
	team, rsp, err := client.Organizations.GetTeam(id)
	if rsp.StatusCode == 304 {
		// return team from cache
		name := d.Get("name").(string)
		description := d.Get("description").(string)
		privacy := d.Get("privacy").(string)
		team = &github.Team{
			Name:        &name,
			Description: &description,
			Privacy:     &privacy,
		}
		err = nil
	}
	return team, err, rsp.Header.Get("ETag")
}

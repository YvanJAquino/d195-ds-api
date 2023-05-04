package main

import (
	"context"
	"sync"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type DataCache struct {
	client *bigquery.Client
	mu     sync.RWMutex
	ready  chan struct{}
	errs   chan error
	data   []Row
}

func NewDataCache(ctx context.Context, projectID string) (*DataCache, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	cache := new(DataCache)
	cache.client = client
	cache.ready = make(chan struct{})
	cache.errs = make(chan error)
	return cache, nil
}

func (c *DataCache) Warmup(ctx context.Context) error {
	c.Build(ctx)
	var err error
	select {
	case <-c.ready:
		err = nil
	case e := <-c.errs:

		err = e
	case <-ctx.Done():
		if e := ctx.Err(); err != nil {

			err = e
		}
	}
	return err

}

func (c *DataCache) Build(ctx context.Context) {
	q := c.client.Query(Query)
	go func() {
		defer close(c.ready)
		defer close(c.errs)
		defer c.mu.Unlock()
		c.mu.Lock()
		it, err := q.Read(ctx)
		if err != nil {
			c.errs <- err
		}
		for {
			var row Row
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				c.errs <- err
			}
			c.data = append(c.data, row)
		}
		c.ready <- struct{}{}
	}()
}

func (c *DataCache) Data() []Row {
	defer c.mu.RUnlock()
	c.mu.RLock()
	return c.data
}

const Query string = "SELECT * FROM `holy-diver-297719.data_science_salaries.curated`"

type Row struct {
	WorkYear          int      `json:"work_year" bigquery:"work_year,nullable"`
	ExperienceLevel   string   `json:"experience_level" bigquery:"experience_level,nullable"`
	EmploymentType    string   `json:"employment_type" bigquery:"employment_type,nullable"`
	JobTitle          string   `json:"job_title" bigquery:"job_title,nullable"`
	Salary            int      `json:"salary" bigquery:"salary,nullable"`
	SalaryCurrency    string   `json:"salary_currency" bigquery:"salary_currency,nullable"`
	SalaryInUsd       int      `json:"salary_in_usd" bigquery:"salary_in_usd,nullable"`
	EmployeeResidence string   `json:"employee_residence" bigquery:"employee_residence,nullable"`
	RemoteRatio       int      `json:"remote_ratio" bigquery:"remote_ratio,nullable"`
	CompanyLocation   string   `json:"company_location" bigquery:"company_location,nullable"`
	CompanySize       string   `json:"company_size" bigquery:"company_size,nullable"`
	TitleTokens       []string `json:"title_tokens" bigquery:"title_tokens,nullable"`
	TitleRole         string   `json:"title_role" bigquery:"title_role,nullable"`
	TitleDomain       string   `json:"title_domain" bigquery:"title_domain,nullable"`
	TitleHonorific    string   `json:"title_honorific" bigquery:"title_honorific,nullable"`
}

type Column string

var (
	WorkYear          Column = "work_year"
	ExperienceLevel   Column = "experience_level"
	CompanyLocation   Column = "company_location"
	EmployeeResidence Column = "employee_residence"
)


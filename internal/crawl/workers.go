package crawl

import (
	"time"

	"github.com/floss-fund/portal/internal/core"
)

func (c *Crawl) dbWorker() {
	var (
		n      = 0
		lastID = 0
	)
	for {
		n++
		items, err := c.db.GetManifestForCrawling(c.opt.ManifestAge, lastID, c.opt.BatchSize)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}

		// No more items. End fetch.
		if len(items) == 0 {
			c.log.Println("no more records to crawl. stopping.")
			break
		}

		for _, i := range items {
			select {
			case c.jobs <- i:
			}
		}

		newID := items[len(items)-1].ID
		c.log.Printf("fetched batch %d of size %d. id %d to %d", n, c.opt.BatchSize, lastID, newID)

		lastID = newID
	}

	// Signal for running workers to quit.
	close(c.jobs)
}

func (c *Crawl) worker() {
loop:
	for {
		select {
		case j, ok := <-c.jobs:
			if !ok {
				break loop
			}

			// Fetch and validate the manifest.
			reCrawl, err := c.IsManifestModified(j.URLobj, j.LastModified)
			if err != nil {
				c.log.Printf("error fetching modified date: %s: %v", j.URL, err)
				continue
			}

			if !reCrawl {
				c.log.Printf("no modification. Skipping: %s", j.URL)
				continue
			}

			// Fetch and validate the manifest.
			status := ""
			m, err := c.FetchManifest(j.URLobj)
			m.ID = j.ID
			if err != nil {
				c.log.Printf("error crawling: %s: %v", j.URL, err)

				// Record the error.
				status, _ = c.db.UpdateManifestCrawlError(j.ID, err.Error(), c.opt.MaxCrawlErrors)
				if c.Callbacks.OnManifestUpdate != nil {
					c.Callbacks.OnManifestUpdate(m, status)
				}

				continue
			}

			// Add it to the database.
			if err := c.db.UpsertManifest(m); err != nil {
				c.log.Printf("error upserting manifest: %s: %v", j.URL, err)
				continue
			}

			if c.Callbacks.OnManifestUpdate != nil {
				c.Callbacks.OnManifestUpdate(m, core.ManifestStatusActive)
			}
		}
	}

	c.wg.Done()
}

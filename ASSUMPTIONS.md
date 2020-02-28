Technical Assumptions

- this is meant to be within an intranet that I've named *.intra.touchtunes.com
  which has a working SSO system that can readily hand out auth tokens. 

- all applications to deploy are versioned with semver, so that understanding updates is more reliable.
  This means their format is strict.

- the CSV is lazy-loaded, to avoid overloading memory; up to n threads are read at a time,
  where n is the number of worker threads set by `-n`.

- The API server is multithreaded, and individual update calls block
  until the receiving device updates (or they return an error if they can't update it),
  so it makes sense to use multiple worker threads to speed the process up.

- I tested two major: parsing input, handling different kinds of API responses,
  and batch updating different numbers of input.
  I did not unit test intentionally triggering the error API responses because
  the challenge said "you do not need to write the API" so I only mocked
  enough of the API to get the 401/409/500/etc responses fed into my code.

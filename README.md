# HubToJo 🐙 → ⚒️: Mirror GitHub repositories to Forgejo

This program creates Forgejo mirrors of the GitHub repositories you specify.

Best run with Docker. Forgejo 11.0.10 or newer is required.

## How to run

```bash
docker run \
    -d \
    --restart=unless-stopped \
    -p 8080:8080 \
    -e FORGEJO_URL="http://forgejo:3000" \
    -e FORGEJO_TOKEN="your Forgejo token" \
    -e GITHUB_USERNAME="your github user" \
    ghcr.io/jdevera/hubtojo:latest
```

The web status page is available on `/`, and the dashboard-friendly JSON stats
endpoint is available on `/stats`.

## Releases

Release images are published to the GitHub Container Registry for Linux AMD64
and ARM64. Publishing a GitHub Release with a semantic version tag builds and
publishes matching image tags:

```bash
gh release create v0.2.0 --generate-notes
```

This publishes `ghcr.io/jdevera/hubtojo:0.2.0`, `:0.2`, and `:latest`. The Git
tag is also embedded in the binary as its version. Starting with `v1.0.0`,
releases also publish a major-only tag such as `:1`. Prereleases publish only
their exact version and do not update floating tags or `:latest`.

## What can be mirrored

- Public Repos: All public repos of the given user, **excluding forks**
- Private Repos: All private repos of the given user
- Forks: All forks of the given user (they are always public)

Each of these groups can be enabled or disabled with the environment variables.

## Parameters

| Parameter                       | Description                                                                      | Mandatory | Default |
|---------------------------------|----------------------------------------------------------------------------------|-----------|---------|
| `FORGEJO_URL`                   | The URL of the Forgejo instance that will mirror the repositories                | Yes       |         |
| `FORGEJO_TOKEN`                 | The token to use when authenticating with the Forgejo API                        | Yes       |         |
| `GITHUB_USERNAME`               | The GitHub username to mirror repositories from                                  | Yes       |         |
| `GITHUB_TOKEN`                  | A GitHub token is required only when working with private repositories           | No        |         |
| `HUBTOJO_MIRROR_PUBLIC_REPOS`   | Set to false or 0 to not mirror public repositories. This does not affect forks. | No        | `true`  |
| `HUBTOJO_MIRROR_PRIVATE_REPOS`  | Set to true or 1 to mirror private repositories                                  | No        | `false` |
| `HUBTOJO_MIRROR_FORKS`          | Set to true or 1 to mirror forks                                                 | No        | `false` |
| `HUBTOJO_DRY_RUN`               | Set to true or 1 to skip the write operations and instead just log them          | No        | `false` |
| `HUBTOJO_NUM_WORKERS`           | The number of concurrent workers to use when mirroring repositories              | No        | `5`     |
| `HUBTOJO_SYNC_INTERVAL`         | The interval in seconds to wait between syncs. Set to 0 to run only once         | No        | `3600`  |
| `HUBTOJO_RUN_TIMEOUT`           | Maximum duration in seconds for startup checks and each synchronization run       | No        | `3600`  |
| `HUBTOJO_WEB_ADDR`              | The address for the status page and stats endpoint                               | No        | `:8080` |


## Custom certificates

If your Forgejo instance is served with a self-signed certificate, or you have a custom CA, `hubtojo` will refuse to connect.

You can provide a custom CA certificate by mounting it in the container. Make sure any additional certificates you want to be trusted are available in the `/usr/local/share/ca-certificates` directory.

With Docker:
```bash
-v /dir/with/certificates:/usr/local/share/ca-certificates:ro
```

With Docker Compose:

```yaml
volumes:
  - /dir/with/certificates:/usr/local/share/ca-certificates:ro
```

The container will automatically trust any certificates in that directory.

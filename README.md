[![Codacy Badge](https://api.codacy.com/project/badge/Grade/c0c64938c02a402f9d8280cf2ecc23d4)](https://www.codacy.com/app/3Blades/core-server?utm_source=github.com&utm_medium=referral&utm_content=3Blades/core-server&utm_campaign=badger)
[![Build Status](https://travis-ci.org/3Blades/core-server.svg?branch=master)](https://travis-ci.org/3Blades/core-server)
[![Docker Hub](https://img.shields.io/badge/docker-ready-blue.svg)](https://registry.hub.docker.com/u/3blades/core-server)
[![codecov](https://codecov.io/gh/3Blades/core-server/branch/master/graph/badge.svg)](https://codecov.io/gh/3Blades/core-server)
[![slack in](https://slack.3blades.io/badge.svg)](https://slack.3blades.io)

# Core server

Core server used for child server types.

Includes the following functionality:

- Authenticates to 3Blades server using JWT's
- SSH remote after_success
- Streaming log capabilities
- Server environment variables
- Cron server
- RESTful endpoint server

Any image could conceivably use core-server functionality. We suggest building custom images with [docker multi-stage builds](https://docs.docker.com/engine/userguide/eng-image/multistage-build/). For example:

```
FROM 3blades/core-server AS core
FROM tensorflow/tensorflow:latest-gpu
COPY --from=core runner /runner
EXPOSE 8080
```

## Copyright and license

Copyright © 2016-2017 3Blades, LLC. All rights reserved, except as follows. Code
is released under the BSD 3.0 license. The README.md file, and files in the
"docs" folder are licensed under the Creative Commons Attribution 4.0
International License under the terms and conditions set forth in the file
"LICENSE.docs". You may obtain a duplicate copy of the same license, titled
CC-BY-SA-4.0, at http://creativecommons.org/licenses/by/4.0/.

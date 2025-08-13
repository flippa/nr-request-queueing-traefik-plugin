# New Relic Request Queueing Traefik Plugin

A Traefik middleware plugin that adds the `X-Request-Start` header to incoming HTTP requests for New Relic request queueing monitoring.

## Overview

This plugin automatically adds a `X-Request-Start` header with the current Unix timestamp in milliseconds to all incoming requests. This header is used by New Relic's Application Performance Monitoring (APM) to measure request queueing time, which helps identify bottlenecks in your application infrastructure.

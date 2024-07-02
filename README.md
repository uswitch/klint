# klint

A tool that listens to changes in Kubernetes resources and runs linting rules against them. Alerts are published
via Slack webhooks to a configurable channel (using an annotation on the object or the object's namespace).

## Table of contents- [klint](#klint)
- [klint](#klint)
  - [Table of contents- klint](#table-of-contents--klint)
  - [Rationale](#rationale)
  - [Building](#building)
  - [Using](#using)
  - [Rules](#rules)
    - [UnsuccessfulExitRule](#unsuccessfulexitrule)
    - [ResourceAnnotationRule](#resourceannotationrule)
    - [ScrapeNeedsPortsRule](#scrapeneedsportsrule)
    - [ValidIAMRoleRule](#validiamrolerule)
    - [RequireCronJobHistoryLimits](#requirecronjobhistorylimits)
  - [Building](#building-1)
  - [Notes](#notes)
  - [License](#license)


## Rationale
We started Klint to help us move more production teams over to our Kubernetes infrastructure. It helps us achieve:

1. Identify and debug erroneous objects
2. Nudge objects in line with our policy as both change over time

For example, we run another tool called [kiam](https://github.com/uswitch/kiam) to integrate with AWS' IAM roles,
allocating each pod its own session-based credentials. On more than one occasion an application team had problems
caused by their roles being spelled incorrectly. Although relatively easy to debug for the Cluster Operations team
it is not a great experience for the application developer. Klint helps us encode such checks and proactively alert
teams when they need to take action.

## Building
To build the exectuable you can use the `go` tool directly:

```
$ go get github.com/uswitch/klint
```

## Using

1. Run klint as a deployment with a single replica on your cluster. 
2. Add an annotation to the namespace, or an object to be monitored: `com.uswitch.alert/slack: <channel>`

As objects change klint will compare them against the rules and post to Slack.

![Alert](alert.png)

## Rules

### UnsuccessfulExitRule
When a Pod exits with a failure code an alert is generated. Additionally, recent log data is retrieved and output
with the message.

If Pods receive `SIGKILL` klint will warn that maybe the `SIGTERM` signal was ignored or that the graceful shutdown
period is too short.

### ResourceAnnotationRule
This ensures that Pods have cpu and memory requests and limits.

### ScrapeNeedsPortsRule
If a Pod is marked as to be scraped via Prometheus (via the `prometheus.io.scrape` annotation) klint will ensure
the Pod also specifies ports. We had instances where applications wanted to be scraped but without the port data
it was unable to figure out what to scrape.

### ValidIAMRoleRule
This rule checks for a valid IAM Role if users specify an IAM role (via the `iam.amazonaws.com/role` annotation). We
had a handful of times where teams had roles with typos or that hadn't been created and it wasn't obvious why
the Pod had no permissions to AWS.

### RequireCronJobHistoryLimits
This currently enforces a relatively low limit insisting that CronJob objects must specify both success and
failure history limits, and that these should both be lower than 10.


## Building

```
$ go build -o bin/klint .
```

## Notes
* *July 2024 -* The `klint` image is now stored in the `uswitch/klint` repository on Quay.
<br>

## License

```
Copyright 2017 uSwitch

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

# flow2granef
Flow2granef is a tool that ingests network security monitoring (NSM) events from different sources (currently Suricata, Zeek and IPFIX flow records) and pushes them into the Dgraph graph database.

The tool is meant to be run as a part of the [GRANEF](https://granef.csirt.muni.cz/) toolkit, reusing its Web and API modules (you have to use modified versions [here](https://github.com/ppochop/granef-analysis-api) and [here](https://github.com/ppochop/granef-analysis-api) as the underlying data model is different). However, you can run it standalone without any frontend, or use the limited browser-based frontend provided by [Dgraph Ratel](https://play.dgraph.io/).

On top of supporting different event sources, the tool is capable of live ingestion. The performance is still being tweaked out so for now, this is only advisable for low bandwidth traffic monitoring.

## Supported event types
The main supported event type are network flows. As the tool is still in the POC stage, there is limited support of application events (in case of IPFIX flows, these are manifested by specific Information Elements in the flow records) scoped to HTTP and DNS.

Only supported event types will be processed by the tool, the rest will be ignored.

## Prerequisites
To use this tool you need to have the following:
  - A reachable Dgraph instance.
  - At least one source of events set up. The events need to be JSON-encoded. Either in a file or in a Kafka topic.

## How to run
All configuration is done via a config file. So to run, you only need to provide the path to the config file:
```
flow2granef --config=config.toml
```

## Configuration
Look into the `config-examples` folder to get a quick understanding.

You need to define each source of events you want to read from. Each source needs:
  - to be uniquely named
  - have the source type provided ("suricata", "zeek", or "ipfixprobe")
  - have the input type provided ("stdin", "file", or "kafka")
  - to have the number of transforming worker threads defined (defaults to 1)
  - to have the input configured with required info

This is done roughly like this (you can define as many of these as you wish, just make sure they have a different name):
```
[sources.uniquename]
transformer = "sourcetype"
input = "inputtype"
workers-num = 8

[sources.uniquename.input-config]
field1 = "some info that configures this particular input type"
```

Additionally, there are some other configuration options you must set:
  - The passive timeout duration (Duration of inactivity after which a flow is considered to be ended). This should correspond to the passive timeout value of the tools you ingest the events from (and should be equal across them).
  - Address of a Dgraph Alpha server.
  - Whether the database should be deleted (reset) before running.
  - Whether to run `flow2granef` in a duplicity check mode.

Example of these options:
```
passive-timeout = "10m"
dgraph-address = "dgraphhost:9080"
reset-dgraph = false
duplicity-check = false
```

### Supported inputs
Other than being able to read events from `stdin`, you can read events from files, or a Kafka topic. Typically, you would use the `file` input for offline ingestion and Kafka for live ingestion.

For now, it is recommended to use only 1 thread for offline ingestion (guarantees that all supported events will be ingested), and multiple threads for live ingestion (can result in skipping some events due to underlying conflicts in Dgraph `upsert` transactions).

A source with the `kafka` input type needs to have the following fields set in the corresponding `input-config` table:
```
[sources.sourcewithkafka.input-config]
bootstrap-servers = "kafkahost1:9092"
group-id = "consumergrouptest"
topic = "eventtopic"
```

A source with the `file` input type needs to have the following fields set in the corresponding `input-config` table:
```
[sources.sourcewithfile.input-config]
path = "path/to/eventfile"
```

## Running with multiple sources
Although you are free to run the tool on several sources, you need to understand what traffic is included in the ingested events as in some situations, this could cause inconsistency in the data.

There are generally two overlaps that can happen:
 - Two sources provide events for the same network traffic (communication between the same hosts) but each source provides a different event type. This is actually a perfectly valid situation supported by `flow2granef` and will result in a consistent state where the application events will be connected to the network flow in the graph database.
 - Two sources provide the same events for the same network traffic. This is not a valid situation. But `flow2granef` offers a mode of operation where the ingested events are checked for this type of duplicity and the user is warned accordingly. This mode is set with the `duplicity-check` field in the config. Note that this mode does not perform any actions with the events (no connection to Dgraph).
# smartctl2prom

smartctl2prom is a tool to convert information presented by a smartctl utility
into Prometheus metrics and make it available for Prometheus in text format
(for node_exporter) or via HTTP protocol. smartctl is a well-knonw utility to
read S.M.A.R.T. data from hard disk drives (HDDs), solid-state drives (SSDs) and
eMMC drives. S.M.A.R.T. (Self-Monitoring, Analysis and Reporting Technology;
often written as SMART) is a monitoring system included in HDDs, SSDs and eMMC
drives.

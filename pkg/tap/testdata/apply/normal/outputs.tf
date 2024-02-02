locals {
  hosts = [
    format("%s.%s.svc.%s", local.resource_name, local.namespace, local.domain_suffix)
  ]

  ports = flatten([
    for c in local.containers : [
      for p in c.ports : try(nonsensitive(p.external), p.external)
      if try(p.external != null, false)
    ]
    if c != null
  ])

  endpoints = length(local.ports) > 0 ? flatten([
    for c in local.hosts : formatlist("%s:%d", c, local.ports)
  ]) : []
}

#
# Orchestration
#

output "context" {
  description = "The input context, a map, which is used for orchestration."
  value       = var.context
}

output "refer" {
  description = "The refer, a map, including hosts, ports and account, which is used for dependencies or collaborations."
  sensitive   = true
  value = {
    schema = "k8s:deployment"
    params = {
      selector  = local.labels
      namespace = local.namespace
      name      = kubernetes_deployment_v1.deployment.metadata[0].name
      hosts     = local.hosts
      ports     = try(nonsensitive(local.ports), local.ports)
      endpoints = try(nonsensitive(local.endpoints), local.endpoints)
    }
  }
}

#
# Reference
#

output "connection" {
  description = "The connection, a string combined host and port, might be a comma separated string or a single string."
  value       = join(",", try(nonsensitive(local.endpoints), local.endpoints))
}

output "address" {
  description = "The address, a string only has host, might be a comma separated string or a single string."
  value       = join(",", local.hosts)
}

output "ports" {
  description = "The port list of the service."
  value       = try(nonsensitive(local.ports), local.ports)
}

#
# Publish
#

data "kubernetes_nodes" "pool" {
  depends_on = [kubernetes_service_v1.service]
}

locals {
  publish_external_hosts = kubernetes_service_v1.service.spec[0].type == "NodePort" ? flatten([
    for n in data.kubernetes_nodes.pool.nodes : [
      for a in n.status[0].addresses : a.address
      if a.type == "ExternalIP"
    ]
    ]) : kubernetes_service_v1.service.spec[0].type == "LoadBalancer" ? flatten([
    for i in kubernetes_service_v1.service.status[0].load_balancer[0].ingress : [
      try(i.hostname != "", false) ? i.hostname : i.ip
    ]
  ]) : []

  publish_internal_hosts = kubernetes_service_v1.service.spec[0].type == "NodePort" ? flatten([
    for n in data.kubernetes_nodes.pool.nodes : [
      for a in n.status[0].addresses : a.address
      if a.type == "InternalIP"
    ]
  ]) : []

  publish_host = length(local.publish_external_hosts) > 0 ? local.publish_external_hosts[0] : length(local.publish_internal_hosts) > 0 ? local.publish_internal_hosts[0] : null
  publish_ports_map = {
    for p in kubernetes_service_v1.service.spec[0].port : p.port => (kubernetes_service_v1.service.spec[0].type == "NodePort" ? p.node_port : p.port)
  }
  publish_endpoints = local.publish_host != null && length(local.publish_ports) > 0 ? {
    for xp in [
      for p in local.publish_ports : p
      if p.schema != null
    ] : format("%d:%d/%s", try(nonsensitive(xp.external), xp.external), try(nonsensitive(xp.internal), xp.internal), try(nonsensitive(xp.schema), xp.schema)) =>
    format("%s://%s:%d", try(nonsensitive(xp.schema), xp.schema), local.publish_host, local.publish_ports_map[try(nonsensitive(xp.external), xp.external)])
  } : {}
}

output "endpoints" {
  description = "The endpoints, a string map, the key is the name, and the value is the URL."
  value       = try(nonsensitive(local.publish_endpoints), local.publish_endpoints)
}

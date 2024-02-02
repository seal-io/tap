terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.23.0"
    }
  }
}

variable "context" {
  description = <<-EOF
Receive contextual information. When Walrus deploys, Walrus will inject specific contextual information into this field.

Examples:
```
context:
  project:
    name: string
    id: string
  environment:
    name: string
    id: string
  resource:
    name: string
    id: string
```
EOF
  type        = map(any)
  default     = {}
}

variable "infrastructure" {
  description = <<-EOF
Specify the infrastructure information for deploying.

Examples:
```
infrastructure:
  namespace: string, optional
  gpu_vendor: string, optional
  domain_suffix: string, optional
  service_type: string, optional
```
EOF
  type = object({
    namespace     = optional(string)
    gpu_vendor    = optional(string, "nvidia.com")
    domain_suffix = optional(string, "cluster.local")
    service_type  = optional(string, "NodePort")
  })
  default = {}
}

variable "deployment" {
  description = <<-EOF
Specify the deployment action, like scaling, scheduling, security and so on.

Examples:
```
deployment:
  timeout: number, optional
  replicas: number, optional
  rolling: 
    max_surge: number, optional          # in fraction, i.e. 0.25, 0.5, 1
    max_unavailable: number, optional    # in fraction, i.e. 0.25, 0.5, 1
  fs_group: number, optional
  sysctls:
  - name: string
    value: string
```
EOF
  type = object({
    timeout  = optional(number, 300)
    replicas = optional(number, 1)
    rolling = optional(object({
      max_surge       = optional(number, 0.25)
      max_unavailable = optional(number, 0.25)
    }))
    fs_group = optional(number)
    sysctls = optional(list(object({
      name  = string
      value = string
    })))
  })
  default = {
    timeout  = 300
    replicas = 1
    rolling = {
      max_surge       = 0.25
      max_unavailable = 0.25
    }
  }
  validation {
    condition     = try(0 < var.deployment.rolling.max_surge && var.deployment.rolling.max_surge <= 1, true)
    error_message = "max_surge must be range from 0.1 to 1"
  }
  validation {
    condition     = try(0 < var.deployment.rolling.max_unavailable && var.deployment.rolling.max_unavailable <= 1, true)
    error_message = "max_surge must be range from 0.1 to 1"
  }
}

variable "containers" {
  description = <<-EOF
Specify the container items to deploy.

Examples:
```
containers:
- profile: init/run
  image: string
  execute:
    working_dir: string, optional
    command: list(string), optional
    args: list(string), optional
    readonly_rootfs: bool, optional
    as_user: number, optional
    as_group: number, optional
    privileged: bool, optional
  resources:
    cpu: number, optional               # in oneCPU, i.e. 0.25, 0.5, 1, 2, 4
    memory: number, optional            # in megabyte
    gpu: number, optional               # in oneGPU, i.e. 1, 2, 4
  envs:
  - name: string
    value: string, optional
    value_refer:
      schema: string
      params: map(any)
  files:
  - path: string
    mode: string, optional
    accept_changed: bool, optional      # accpet changed
    content: string, optional
    content_refer:
      schema: string
      params: map(any)
  mounts:
  - path: string
    readonly: bool, optional
    subpath: string, optional
    volume: string, optional            # shared between containers if named, otherwise exclusively by this container
    volume_refer:
      schema: string
      params: map(any)
  ports:
  - internal: number
    external: number, optional
    protocol: tcp/udp
    schema: string, optional
  checks:
  - type: execute/tcp/http/https
    delay: number, optional
    interval: number, optional
    timeout: number, optional
    retries: number, optional
    teardown: bool, optional
    execute:
      command: list(string)
    tcp:
      port: number
    http:
      port: number
      headers: map(string), optional
      path: string, optional
    https:
      port: number
      headers: map(string), optional
      path: string, optional
```
EOF
  type = list(object({
    profile = optional(string, "run")
    image   = string
    execute = optional(object({
      working_dir     = optional(string)
      command         = optional(list(string))
      args            = optional(list(string))
      readonly_rootfs = optional(bool, false)
      as_user         = optional(number)
      as_group        = optional(number)
      privileged      = optional(bool, false)
    }))
    resources = optional(object({
      cpu    = optional(number, 0.25)
      memory = optional(number, 256)
      gpu    = optional(number, 0)
    }))
    envs = optional(list(object({
      name  = string
      value = optional(string)
      value_refer = optional(object({
        schema = string
        params = map(any)
      }))
    })))
    files = optional(list(object({
      path           = string
      mode           = optional(string, "0644")
      accept_changed = optional(bool, false)
      content        = optional(string)
      content_refer = optional(object({
        schema = string
        params = map(any)
      }))
    })))
    mounts = optional(list(object({
      path     = string
      readonly = optional(bool, false)
      subpath  = optional(string)
      volume   = optional(string)
      volume_refer = optional(object({
        schema = string
        params = map(any)
      }))
    })))
    ports = optional(list(object({
      internal = number
      external = optional(number)
      protocol = optional(string, "tcp")
      schema   = optional(string)
    })))
    checks = optional(list(object({
      type     = string
      delay    = optional(number, 0)
      interval = optional(number, 10)
      timeout  = optional(number, 1)
      retries  = optional(number, 1)
      teardown = optional(bool, false)
      execute = optional(object({
        command = list(string)
      }))
      tcp = optional(object({
        port = number
      }))
      http = optional(object({
        port    = number
        headers = optional(map(string))
        path    = optional(string, "/")
      }))
      https = optional(object({
        port    = number
        headers = optional(map(string))
        path    = optional(string, "/")
      }))
    })))
  }))
  validation {
    condition     = length(var.containers) > 0
    error_message = "containers must be at least one"
  }
  validation {
    condition     = alltrue([for c in var.containers : try(c.profile == "" || contains(["init", "run"], c.profile), true)])
    error_message = "profile must be init or run"
  }
  validation {
    condition = alltrue(flatten([
      for c in var.containers : [
        for p in try(c.ports != null ? c.ports : [], []) : try(0 < p.internal && p.internal < 65536, true) && try(0 < p.external && p.external < 65536, true)
      ]
    ]))
    error_message = "port must be range from 1 to 65535"
  }
}

locals {
  project_name     = coalesce(try(var.context["project"]["name"], null), "default")
  project_id       = coalesce(try(var.context["project"]["id"], null), "default_id")
  environment_name = coalesce(try(var.context["environment"]["name"], null), "test")
  environment_id   = coalesce(try(var.context["environment"]["id"], null), "test_id")
  resource_name    = coalesce(try(var.context["resource"]["name"], null), "example")
  resource_id      = coalesce(try(var.context["resource"]["id"], null), "example_id")
  namespace = coalesce(try(var.infrastructure.namespace, ""), join("-", [
    local.project_name, local.environment_name
  ]))
  gpu_vendor    = coalesce(try(var.infrastructure.gpu_vendor, ""), "nvdia.com")
  domain_suffix = coalesce(var.infrastructure.domain_suffix, "cluster.local")
  annotations = {
    "walrus.seal.io/project-id"     = local.project_id
    "walrus.seal.io/environment-id" = local.environment_id
    "walrus.seal.io/resource-id"    = local.resource_id
  }
  labels = {
    "walrus.seal.io/catalog-name"     = "terraform-kubernetes-containerservice"
    "walrus.seal.io/project-name"     = local.project_name
    "walrus.seal.io/environment-name" = local.environment_name
    "walrus.seal.io/resource-name"    = local.resource_name
  }
  wellknown_env_schemas    = ["k8s:secret"]
  wellknown_file_schemas   = ["k8s:secret", "k8s:configmap"]
  wellknown_mount_schemas  = ["k8s:secret", "k8s:configmap", "k8s:persistentvolumeclaim"]
  wellknown_port_protocols = ["TCP", "UDP"]
  internal_port_container_index_map = {
    for ip, cis in merge(flatten([
      for i, c in var.containers : [{
        for p in try(c.ports != null ? c.ports : [], []) : p.internal => i...
        if p != null
      }]
    ])...) : ip => cis[0]
  }
  containers = [
    for i, c in var.containers : merge(c, {
      name = format("%s-%d-%s", coalesce(c.profile, "run"), i, basename(split(":", c.image)[0]))
      envs = [
        for xe in [
          for e in(c.envs != null ? c.envs : []) : e
          if e != null && try(!(e.value != null && e.value_refer != null) && !(e.value == null && e.value_refer == null), false)
        ] : xe
        if xe.value_refer == null || (try(contains(local.wellknown_env_schemas, xe.value_refer.schema), false) && try(lookup(xe.value_refer.params, "name", null) != null, false) && try(lookup(xe.value_refer.params, "key", null) != null, false))
      ]
      files = [
        for xf in [
          for f in(c.files != null ? c.files : []) : f
          if f != null && try(!(f.content != null && f.content_refer != null) && !(f.content == null && f.content_refer == null), false)
        ] : xf
        if xf.content_refer == null || (try(contains(local.wellknown_file_schemas, xf.content_refer.schema), false) && try(lookup(xf.content_refer.params, "name", null) != null, false) && try(lookup(xf.content_refer.params, "key", null) != null, false))
      ]
      mounts = [
        for xm in [
          for m in(c.mounts != null ? c.mounts : []) : m
          if m != null && try(!(m.volume != null && m.volume_refer != null), false)
        ] : xm
        if xm.volume_refer == null || (try(contains(local.wellknown_mount_schemas, xm.volume_refer.schema), false) && try(lookup(xm.volume_refer.params, "name", null) != null, false))
      ]
      ports = [
        for xp in [
          for _, ps in {
            for p in(c.ports != null ? c.ports : []) : p.internal => {
              internal = p.internal
              external = p.external
              protocol = p.protocol == null ? "TCP" : upper(p.protocol)
              schema   = p.schema == null ? (contains([80, 8080], p.internal) ? "http" : (contains([443, 8443], p.internal) ? "https" : null)) : lower(p.schema)
            }...
            if p != null
          } : ps[length(ps) - 1]
          if local.internal_port_container_index_map[ps[length(ps) - 1].internal] == i
        ] : xp
        if try(contains(local.wellknown_port_protocols, xp.protocol), true)
      ]
      checks = [
        for ck in(c.checks != null ? c.checks : []) : ck
        if try(lookup(ck, ck.type, null) != null, false)
      ]
    })
    if c != null
  ]
  container_ephemeral_envs_map = {
    for c in local.containers : c.name => [
      for e in c.envs : e
      if try(e.value_refer == null, false)
    ]
    if c != null
  }
  container_refer_envs_map = {
    for c in local.containers : c.name => [
      for e in c.envs : e
      if try(e.value_refer != null, false)
    ]
    if c != null
  }
  container_ephemeral_files_map = {
    for c in local.containers : c.name => [
      for f in c.files : merge(f, {
        name = format("eph-f-%s-%s", c.name, md5(f.path))
      })
      if try(f.content_refer == null, false)
    ]
    if c != null
  }
  container_refer_files_map = {
    for c in local.containers : c.name => [
      for f in c.files : merge(f, {
        name = format("ref-f-%s-%s", c.name, md5(jsonencode(f.content_refer)))
      })
      if try(f.content_refer != null, false)
    ]
    if c != null
  }
  container_ephemeral_mounts_map = {
    for c in local.containers : c.name => [
      for m in c.mounts : merge(m, {
        name = format("eph-m-%s", try(m.volume == null || m.volume == "", true) ? md5(join("/", [c.name, m.path])) : md5(m.volume))
      })
      if try(m.volume_refer == null, false)
    ]
    if c != null
  }
  container_refer_mounts_map = {
    for c in local.containers : c.name => [
      for m in c.mounts : merge(m, {
        name = format("ref-m-%s", md5(jsonencode(m.volume_refer)))
      })
      if try(m.volume_refer != null, false)
    ]
    if c != null
  }
  container_internal_ports_map = {
    for c in local.containers : c.name => [
      for p in c.ports : merge(p, {
        name = lower(format("%s-%d", p.protocol, p.internal))
      })
      if p != null
    ]
    if c != null
  }
  init_containers = [
    for c in local.containers : c
    if c != null && try(c.profile == "init", false)
  ]
  run_containers = [
    for c in local.containers : c
    if c != null && try(c.profile == "" || c.profile == "run", true)
  ]
  ephemeral_files = flatten([
    for _, fs in local.container_ephemeral_files_map : fs
  ])
  refer_files = flatten([
    for _, fs in local.container_refer_files_map : fs
  ])
  ephemeral_mounts = [
    for _, v in {
      for m in flatten([
        for _, ms in local.container_ephemeral_mounts_map : ms
      ]) : m.name => m...
    } : v[0]
  ]
  refer_mounts = [
    for _, v in {
      for m in flatten([
        for _, ms in local.container_refer_mounts_map : ms
      ]) : m.name => m...
    } : v[0]
  ]
  ephemeral_files_map = {
    for f in local.ephemeral_files : f.name => f
  }
  downward_annotations = {
    WALRUS_PROJECT_ID     = "walrus.seal.io/project-id"
    WALRUS_ENVIRONMENT_ID = "walrus.seal.io/environment-id"
    WALRUS_RESOURCE_ID    = "walrus.seal.io/resource-id"
  }
  downward_labels = {
    WALRUS_PROJECT_NAME     = "walrus.seal.io/project-name"
    WALRUS_ENVIRONMENT_NAME = "walrus.seal.io/environment-name"
    WALRUS_RESOURCE_NAME    = "walrus.seal.io/resource-name"
  }
  run_containers_mapping_checks_map = {
    for n, cks in {
      for c in local.run_containers : c.name => {
        startup = [
          for ck in c.checks : ck
          if try(ck.delay > 0 && ck.teardown, false)
        ]
        readiness = [
          for ck in c.checks : ck
          if try(!ck.teardown, false)
        ]
        liveness = [
          for ck in c.checks : ck
          if try(ck.teardown, false)
        ]
      }
      } : n => merge(cks, {
        startup   = try(slice(cks.startup, 0, 1), [])
        readiness = try(slice(cks.readiness, 0, 1), [])
        liveness  = try(slice(cks.liveness, 0, 1), [])
    })
  }
  service_type = try(coalesce(var.infrastructure.service_type, "NodePort"), "NodePort")
  publish_ports = flatten([
    for c in local.containers : [
      for p in c.ports : p
      if try(p.external != null, false)
    ]
    if c != null
  ])
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

resource "kubernetes_config_map_v1" "ephemeral_files" {
  for_each = toset(keys(try(nonsensitive(local.ephemeral_files_map), local.ephemeral_files_map)))
  data = {
    content = local.ephemeral_files_map[each.key].content
  }
  metadata {
    namespace   = local.namespace
    name        = each.key
    annotations = local.annotations
    labels      = local.labels
  }
}

resource "kubernetes_deployment_v1" "deployment" {
  wait_for_rollout = false
  metadata {
    namespace     = local.namespace
    generate_name = format("%s-", local.resource_name)
    annotations   = local.annotations
    labels        = local.labels
  }
  spec {
    min_ready_seconds         = 0
    revision_history_limit    = 3
    progress_deadline_seconds = try(var.deployment.timeout != null && var.deployment.timeout > 0, false) ? var.deployment.timeout : null
    replicas                  = var.deployment.replicas
    strategy {
      type = "RollingUpdate"
      rolling_update {
        max_surge       = format("%d%%", try(var.deployment.rolling.max_surge, 0.25) * 100)
        max_unavailable = format("%d%%", try(var.deployment.rolling.max_unavailable, 0.25) * 100)
      }
    }
    selector {
      match_labels = local.labels
    }
    template {
      metadata {
        annotations = local.annotations
        labels      = local.labels
      }
      spec {
        automount_service_account_token = false
        restart_policy                  = "Always"
        dynamic "security_context" {
          for_each = try(length(var.deployment.sysctls), 0) > 0 || try(var.deployment.fs_group != null, false) ? [{}] : []
          content {
            fs_group = try(var.deployment.fs_group, null)
            dynamic "sysctl" {
              for_each = try(var.deployment.sysctls != null, false) ? try(nonsensitive(var.deployment.sysctls), var.deployment.sysctls) : []
              content {
                name  = sysctl.value.name
                value = sysctl.value.value
              }
            }
          }
        }
        dynamic "volume" {
          for_each = try(nonsensitive(local.ephemeral_files), local.ephemeral_files)
          content {
            name = volume.value.name
            config_map {
              default_mode = volume.value.mode
              name         = volume.value.name
              items {
                key  = "content"
                path = basename(volume.value.path)
              }
            }
          }
        }
        dynamic "volume" {
          for_each = try(nonsensitive(local.refer_files), local.refer_files)
          content {
            name = volume.value.name
            dynamic "config_map" {
              for_each = volume.value.content_refer.schema == "k8s:configmap" ? [try(nonsensitive(volume.value), volume.value)] : []
              content {
                default_mode = config_map.value.mode
                name         = config_map.value.content_refer.params.name
                optional     = try(lookup(config_map.value.volume_refer.params, "optional", null), null)
                items {
                  key  = config_map.value.content_refer.params.key
                  path = basename(config_map.value.path)
                }
              }
            }
            dynamic "secret" {
              for_each = volume.value.content_refer.schema == "k8s:secret" ? [try(nonsensitive(volume.value), volume.value)] : []
              content {
                default_mode = secret.value.mode
                secret_name  = secret.value.content_refer.params.name
                optional     = try(lookup(secret.value.volume_refer.params, "optional", null), null)
                items {
                  key  = secret.value.content_refer.params.key
                  path = basename(secret.value.path)
                }
              }
            }
          }
        }
        dynamic "volume" {
          for_each = try(nonsensitive(local.ephemeral_mounts), local.ephemeral_mounts)
          content {
            name = volume.value.name
            empty_dir {
            }
          }
        }
        dynamic "volume" {
          for_each = try(nonsensitive(local.refer_mounts), local.refer_mounts)
          content {
            name = volume.value.name
            dynamic "config_map" {
              for_each = volume.value.volume_refer.schema == "k8s:configmap" ? [try(nonsensitive(volume.value), volume.value)] : []
              content {
                default_mode = try(lookup(config_map.value.volume_refer.params, "mode", null), null)
                name         = config_map.value.volume_refer.params.name
                optional     = try(lookup(config_map.value.volume_refer.params, "optional", null), null)
              }
            }
            dynamic "secret" {
              for_each = volume.value.volume_refer.schema == "k8s:secret" ? [try(nonsensitive(volume.value), volume.value)] : []
              content {
                default_mode = try(lookup(secret.value.volume_refer.params, "mode", null), null)
                secret_name  = secret.value.volume_refer.params.name
                optional     = try(lookup(secret.value.volume_refer.params, "optional", null), null)
              }
            }
            dynamic "persistent_volume_claim" {
              for_each = volume.value.volume_refer.schema == "k8s:persistentvolumeclaim" ? [try(nonsensitive(volume.value), volume.value)] : []
              content {
                read_only  = try(lookup(persistent_volume_claim.value.volume_refer.params, "readonly", null), false)
                claim_name = persistent_volume_claim.value.volume_refer.params.name
              }
            }
          }
        }
        dynamic "init_container" {
          for_each = try(nonsensitive(local.init_containers), local.init_containers)
          content {
            name              = init_container.value.name
            image             = init_container.value.image
            image_pull_policy = "IfNotPresent"
            working_dir       = try(init_container.value.execute.working_dir, null)
            command           = try(init_container.value.execute.command, null)
            args              = try(init_container.value.execute.args, null)
            security_context {
              read_only_root_filesystem = try(init_container.value.execute.readonly_rootfs, false)
              run_as_user               = try(init_container.value.execute.as_user, null)
              run_as_group              = try(init_container.value.execute.as_group, null)
              privileged                = try(init_container.value.execute.privileged, null)
            }
            dynamic "resources" {
              for_each = init_container.value.resources != null ? try([nonsensitive(init_container.value.resources)], [init_container.value.resources]) : []
              content {
                requests = {
                  for k, v in resources.value : "%{if k == "gpu"}${local.gpu_vendor}/%{endif}${k}" => "%{if k == "memory"}${v}Mi%{else}${v}%{endif}"
                  if try(v != null && v > 0, false)
                }
                limits = {
                  for k, v in resources.value : "%{if k == "gpu"}${local.gpu_vendor}/%{endif}${k}" => "%{if k == "memory"}${v}Mi%{else}${v}%{endif}"
                  if try(v != null && v > 0, false) && k != "cpu"
                }
              }
            }
            dynamic "env" {
              for_each = local.container_ephemeral_envs_map[init_container.value.name] != null ? try(nonsensitive(local.container_ephemeral_envs_map[init_container.value.name]), local.container_ephemeral_envs_map[init_container.value.name]) : []
              content {
                name  = env.value.name
                value = env.value.value
              }
            }
            dynamic "env" {
              for_each = local.container_refer_envs_map[init_container.value.name] != null ? try(nonsensitive(local.container_refer_envs_map[init_container.value.name]), local.container_refer_envs_map[init_container.value.name]) : []
              content {
                name = env.value.name
                value_from {
                  secret_key_ref {
                    name = env.value.value_refer.params.name
                    key  = env.value.value_refer.params.key
                  }
                }
              }
            }
            dynamic "env" {
              for_each = local.downward_annotations
              content {
                name = env.key
                value_from {
                  field_ref {
                    field_path = format("metadata.annotations['%s']", env.value)
                  }
                }
              }
            }
            dynamic "env" {
              for_each = local.downward_labels
              content {
                name = env.key
                value_from {
                  field_ref {
                    field_path = format("metadata.labels['%s']", env.value)
                  }
                }
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_ephemeral_files_map[init_container.value.name] != null ? try(nonsensitive(local.container_ephemeral_files_map[init_container.value.name]), local.container_ephemeral_files_map[init_container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = try(volume_mount.value.accept_changed, false) ? dirname(volume_mount.value.path) : volume_mount.value.path
                sub_path   = try(volume_mount.value.accept_changed, false) ? null : basename(volume_mount.value.path)
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_refer_files_map[init_container.value.name] != null ? try(nonsensitive(local.container_refer_files_map[init_container.value.name]), local.container_refer_files_map[init_container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = try(volume_mount.value.accept_changed, false) ? dirname(volume_mount.value.path) : volume_mount.value.path
                sub_path   = try(volume_mount.value.accept_changed, false) ? null : basename(volume_mount.value.path)
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_ephemeral_mounts_map[init_container.value.name] != null ? try(nonsensitive(local.container_ephemeral_mounts_map[init_container.value.name]), local.container_ephemeral_mounts_map[init_container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = volume_mount.value.path
                read_only  = try(volume_mount.value.readonly, null)
                sub_path   = try(volume_mount.value.subpath, null)
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_refer_mounts_map[init_container.value.name] != null ? try(nonsensitive(local.container_refer_mounts_map[init_container.value.name]), local.container_refer_mounts_map[init_container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = volume_mount.value.path
                read_only  = try(volume_mount.value.readonly, null)
                sub_path   = try(volume_mount.value.subpath, null)
              }
            }
          }
        }
        dynamic "container" {
          for_each = try(nonsensitive(local.run_containers), local.run_containers)
          content {
            name              = container.value.name
            image             = container.value.image
            image_pull_policy = "IfNotPresent"
            working_dir       = try(container.value.execute.working_dir, null)
            command           = try(container.value.execute.command, null)
            args              = try(container.value.execute.args, null)
            security_context {
              read_only_root_filesystem = try(container.value.execute.readonly_rootfs, false)
              run_as_user               = try(container.value.execute.as_user, null)
              run_as_group              = try(container.value.execute.as_group, null)
              privileged                = try(container.value.execute.privileged, null)
            }
            dynamic "resources" {
              for_each = container.value.resources != null ? try([nonsensitive(container.value.resources)], [container.value.resources]) : []
              content {
                requests = {
                  for k, v in resources.value : "%{if k == "gpu"}${local.gpu_vendor}/%{endif}${k}" => "%{if k == "memory"}${v}Mi%{else}${v}%{endif}"
                  if try(v != null && v > 0, false)
                }
                limits = {
                  for k, v in resources.value : "%{if k == "gpu"}${local.gpu_vendor}/%{endif}${k}" => "%{if k == "memory"}${v}Mi%{else}${v}%{endif}"
                  if try(v != null && v > 0, false) && k != "cpu"
                }
              }
            }
            dynamic "env" {
              for_each = local.container_ephemeral_envs_map[container.value.name] != null ? try(nonsensitive(local.container_ephemeral_envs_map[container.value.name]), local.container_ephemeral_envs_map[container.value.name]) : []
              content {
                name  = env.value.name
                value = env.value.value
              }
            }
            dynamic "env" {
              for_each = local.container_refer_envs_map[container.value.name] != null ? try(nonsensitive(local.container_refer_envs_map[container.value.name]), local.container_refer_envs_map[container.value.name]) : []
              content {
                name = env.value.name
                value_from {
                  secret_key_ref {
                    name = env.value.value_refer.params.name
                    key  = env.value.value_refer.params.key
                  }
                }
              }
            }
            dynamic "env" {
              for_each = local.downward_annotations
              content {
                name = env.key
                value_from {
                  field_ref {
                    field_path = format("metadata.annotations['%s']", env.value)
                  }
                }
              }
            }
            dynamic "env" {
              for_each = local.downward_labels
              content {
                name = env.key
                value_from {
                  field_ref {
                    field_path = format("metadata.labels['%s']", env.value)
                  }
                }
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_ephemeral_files_map[container.value.name] != null ? try(nonsensitive(local.container_ephemeral_files_map[container.value.name]), local.container_ephemeral_files_map[container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = try(volume_mount.value.accept_changed, false) ? dirname(volume_mount.value.path) : volume_mount.value.path
                sub_path   = try(volume_mount.value.accept_changed, false) ? null : basename(volume_mount.value.path)
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_refer_files_map[container.value.name] != null ? try(nonsensitive(local.container_refer_files_map[container.value.name]), local.container_refer_files_map[container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = try(volume_mount.value.accept_changed, false) ? dirname(volume_mount.value.path) : volume_mount.value.path
                sub_path   = try(volume_mount.value.accept_changed, false) ? null : basename(volume_mount.value.path)
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_ephemeral_mounts_map[container.value.name] != null ? try(nonsensitive(local.container_ephemeral_mounts_map[container.value.name]), local.container_ephemeral_mounts_map[container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = volume_mount.value.path
                read_only  = try(volume_mount.value.readonly, null)
                sub_path   = try(volume_mount.value.subpath, null)
              }
            }
            dynamic "volume_mount" {
              for_each = local.container_refer_mounts_map[container.value.name] != null ? try(nonsensitive(local.container_refer_mounts_map[container.value.name]), local.container_refer_mounts_map[container.value.name]) : []
              content {
                name       = volume_mount.value.name
                mount_path = volume_mount.value.path
                read_only  = try(volume_mount.value.readonly, null)
                sub_path   = try(volume_mount.value.subpath, null)
              }
            }
            dynamic "port" {
              for_each = local.container_internal_ports_map[container.value.name] != null ? try(nonsensitive(local.container_internal_ports_map[container.value.name]), local.container_internal_ports_map[container.value.name]) : []
              content {
                name           = port.value.name
                protocol       = port.value.protocol
                container_port = port.value.internal
              }
            }
            dynamic "startup_probe" {
              for_each = try(nonsensitive(local.run_containers_mapping_checks_map[container.value.name].startup), local.run_containers_mapping_checks_map[container.value.name].startup)
              content {
                initial_delay_seconds = startup_probe.value.delay
                period_seconds        = startup_probe.value.interval
                timeout_seconds       = startup_probe.value.timeout
                failure_threshold     = startup_probe.value.retries
                dynamic "exec" {
                  for_each = startup_probe.value.type == "execute" ? [try(nonsensitive(startup_probe.value.execute), startup_probe.value.execute)] : []
                  content {
                    command = exec.value.command
                  }
                }
                dynamic "tcp_socket" {
                  for_each = startup_probe.value.type == "tcp" ? [try(nonsensitive(startup_probe.value.tcp), startup_probe.value.tcp)] : []
                  content {
                    port = tcp_socket.value.port
                  }
                }
                dynamic "http_get" {
                  for_each = startup_probe.value.type == "http" ? [try(nonsensitive(startup_probe.value.http), startup_probe.value.http)] : []
                  content {
                    port   = http_get.value.port
                    path   = http_get.value.path
                    scheme = "HTTP"
                    dynamic "http_header" {
                      for_each = try(http_get.value.headers != null, false) ? try(nonsensitive(http_get.value.headers), http_get.value.headers) : {}
                      content {
                        name  = http_header.key
                        value = http_header.value
                      }
                    }
                  }
                }
                dynamic "http_get" {
                  for_each = startup_probe.value.type == "https" ? [try(nonsensitive(startup_probe.value.https), startup_probe.value.https)] : []
                  content {
                    port   = http_get.value.port
                    path   = http_get.value.path
                    scheme = "HTTPS"
                    dynamic "http_header" {
                      for_each = try(http_get.value.headers != null, false) ? try(nonsensitive(http_get.value.headers), http_get.value.headers) : {}
                      content {
                        name  = http_header.key
                        value = http_header.value
                      }
                    }
                  }
                }
              }
            }
            dynamic "readiness_probe" {
              for_each = try(nonsensitive(local.run_containers_mapping_checks_map[container.value.name].readiness), local.run_containers_mapping_checks_map[container.value.name].readiness)
              content {
                initial_delay_seconds = readiness_probe.value.delay
                period_seconds        = readiness_probe.value.interval
                timeout_seconds       = readiness_probe.value.timeout
                failure_threshold     = readiness_probe.value.retries
                dynamic "exec" {
                  for_each = readiness_probe.value.type == "execute" ? [try(nonsensitive(readiness_probe.value.execute), readiness_probe.value.execute)] : []
                  content {
                    command = exec.value.command
                  }
                }
                dynamic "tcp_socket" {
                  for_each = readiness_probe.value.type == "tcp" ? [try(nonsensitive(readiness_probe.value.tcp), readiness_probe.value.tcp)] : []
                  content {
                    port = tcp_socket.value.port
                  }
                }
                dynamic "http_get" {
                  for_each = readiness_probe.value.type == "http" ? [try(nonsensitive(readiness_probe.value.http), readiness_probe.value.http)] : []
                  content {
                    port   = http_get.value.port
                    path   = http_get.value.path
                    scheme = "HTTP"
                    dynamic "http_header" {
                      for_each = try(http_get.value.headers != null, false) ? try(nonsensitive(http_get.value.headers), http_get.value.headers) : {}
                      content {
                        name  = http_header.key
                        value = http_header.value
                      }
                    }
                  }
                }
                dynamic "http_get" {
                  for_each = readiness_probe.value.type == "https" ? [try(nonsensitive(readiness_probe.value.https), readiness_probe.value.https)] : []
                  content {
                    port   = http_get.value.port
                    path   = http_get.value.path
                    scheme = "HTTPS"
                    dynamic "http_header" {
                      for_each = try(http_get.value.headers != null, false) ? try(nonsensitive(http_get.value.headers), http_get.value.headers) : {}
                      content {
                        name  = http_header.key
                        value = http_header.value
                      }
                    }
                  }
                }
              }
            }
            dynamic "liveness_probe" {
              for_each = try(nonsensitive(local.run_containers_mapping_checks_map[container.value.name].liveness), local.run_containers_mapping_checks_map[container.value.name].liveness)
              content {
                period_seconds    = liveness_probe.value.interval
                timeout_seconds   = liveness_probe.value.timeout
                failure_threshold = liveness_probe.value.retries
                dynamic "exec" {
                  for_each = liveness_probe.value.type == "execute" ? [try(nonsensitive(liveness_probe.value.execute), liveness_probe.value.execute)] : []
                  content {
                    command = exec.value.command
                  }
                }
                dynamic "tcp_socket" {
                  for_each = liveness_probe.value.type == "tcp" ? [try(nonsensitive(liveness_probe.value.tcp), liveness_probe.value.tcp)] : []
                  content {
                    port = tcp_socket.value.port
                  }
                }
                dynamic "http_get" {
                  for_each = liveness_probe.value.type == "http" ? [try(nonsensitive(liveness_probe.value.http), liveness_probe.value.http)] : []
                  content {
                    port   = http_get.value.port
                    path   = http_get.value.path
                    scheme = "HTTP"
                    dynamic "http_header" {
                      for_each = try(http_get.value.headers != null, false) ? try(nonsensitive(http_get.value.headers), http_get.value.headers) : {}
                      content {
                        name  = http_header.key
                        value = http_header.value
                      }
                    }
                  }
                }
                dynamic "http_get" {
                  for_each = liveness_probe.value.type == "https" ? [try(nonsensitive(liveness_probe.value.https), liveness_probe.value.https)] : []
                  content {
                    port   = http_get.value.port
                    path   = http_get.value.path
                    scheme = "HTTPS"
                    dynamic "http_header" {
                      for_each = try(http_get.value.headers != null, false) ? try(nonsensitive(http_get.value.headers), http_get.value.headers) : {}
                      content {
                        name  = http_header.key
                        value = http_header.value
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}

resource "terraform_data" "replacement" {
  input = sha256(jsonencode({
    is_loadbalancer   = local.service_type == "LoadBalancer"
    has_publish_ports = length(try(nonsensitive(local.publish_ports), local.publish_ports)) > 0
  }))
}

resource "kubernetes_service_v1" "service" {
  wait_for_load_balancer = local.service_type == "LoadBalancer"
  metadata {
    namespace   = local.namespace
    name        = local.resource_name
    annotations = local.annotations
    labels      = local.labels
  }
  spec {
    selector         = local.labels
    type             = length(local.publish_ports) > 0 ? local.service_type : "ClusterIP"
    session_affinity = length(local.publish_ports) > 0 && local.service_type == "ClientIP" ? "ClientIP" : "None"
    cluster_ip       = length(local.publish_ports) > 0 ? null : "None"
    dynamic "port" {
      for_each = try(nonsensitive(local.publish_ports), local.publish_ports)
      content {
        name        = lower(format("%s-%d", port.value.protocol, port.value.external))
        port        = port.value.external
        target_port = port.value.internal
        protocol    = port.value.protocol
      }
    }
  }
  lifecycle {
    replace_triggered_by = [terraform_data.replacement]
  }
}

data "kubernetes_nodes" "pool" {
  depends_on = [kubernetes_service_v1.service]
}

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

output "endpoints" {
  description = "The endpoints, a string map, the key is the name, and the value is the URL."
  value       = try(nonsensitive(local.publish_endpoints), local.publish_endpoints)
}

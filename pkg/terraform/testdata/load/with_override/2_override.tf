resource "kubernetes_deployment_v1" "deploy" {
  metadata {
    name      = "override2"
    namespace = "override2"
  }
}

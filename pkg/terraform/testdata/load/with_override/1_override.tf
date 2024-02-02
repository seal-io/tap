resource "kubernetes_deployment_v1" "deploy" {
  metadata {
    name      = "override1"
    namespace = "override1"
  }
}

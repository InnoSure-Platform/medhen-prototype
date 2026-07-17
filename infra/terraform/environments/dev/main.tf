# Core Namespaces
resource "kubernetes_namespace" "iam" {
  metadata {
    name = "medhen-iam"
  }
}

resource "kubernetes_namespace" "core_services" {
  metadata {
    name = "medhen-core"
  }
}

# Deploy Keycloak via Helm (Bitnami Chart)
resource "helm_release" "keycloak" {
  name       = "keycloak"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "keycloak"
  namespace  = kubernetes_namespace.iam.metadata[0].name
  version    = "21.4.2"

  set {
    name  = "auth.adminUser"
    value = "admin"
  }

  set {
    name  = "auth.adminPassword"
    value = "admin123" # Must be injected via secret in real prod
  }

  # Enable Postgres Backend
  set {
    name  = "postgresql.enabled"
    value = "true"
  }
  
  # For local/dev testing without Ingress constraints
  set {
    name  = "proxy"
    value = "edge"
  }
}

# In a real environment, we would also deploy OPA Gatekeeper or centralized config maps here.

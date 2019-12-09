variable "gke_cluster_num_nodes" {
  default = 3
  description = ""
  type    = "string"
}

variable "gke_cluster_node_machine_type" {
  default = "n1-standard-4"
  description = ""
  type    = "string"
}

variable "eirini_system_namespace" {
  default = "cf-system"
  description = ""
  type    = "string"
}

variable "eirini_service_account_name" {
  default = "opi-service-account"
  description = ""
  type    = "string"
}

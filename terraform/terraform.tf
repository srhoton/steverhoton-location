terraform {
  backend "s3" {
    bucket = "srhoton-tfstate"
    key    = "location-lambda/terraform.tfstate"
    region = "us-east-1"
  }
}
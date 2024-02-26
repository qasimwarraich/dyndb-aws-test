module "lambda-function" {
    source = "./modules/terraform-aws-lambda"
}

module "dynamo-db" {
    source = "./modules/terraform-aws-dynamodb"
}

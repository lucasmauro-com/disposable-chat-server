resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "disposable-chat-server-redis"
  engine               = "redis"
  node_type            = "cache.t2.small"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379
  apply_immediately    = true
  subnet_group_name    = aws_elasticache_subnet_group.disposable_chat.name
  security_group_ids   = [aws_security_group.disposable_chat.id]
}

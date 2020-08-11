# API GatewayへPOSTする例
curl -X POST -H "Content-Type: application/json" -d '{"user_id": "12345", "user_name":"テストユーザー", "age":"25", "send_push":"true", "send_email": "false"}' https://your-endopoint

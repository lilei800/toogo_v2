#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
HotGo 强密码生成工具
生成所有需要的密码和密钥
"""

import secrets
import string
from datetime import datetime

def generate_password(length=16):
    """生成强密码"""
    chars = string.ascii_letters + string.digits + '!@#$%^&*+-=?'
    return ''.join(secrets.choice(chars) for _ in range(length))

def generate_hex_key(length=32):
    """生成十六进制密钥"""
    return secrets.token_hex(length)

def main():
    print("=" * 60)
    print("HotGo 强密码生成工具")
    print("=" * 60)
    print()
    print(f"生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print()
    
    # 1. PostgreSQL密码
    pg_password = generate_password(16)
    print("1. PostgreSQL密码 (16字符):")
    print(f"   {pg_password}")
    print()
    
    # 2. Redis密码
    redis_password = generate_password(16)
    print("2. Redis密码 (16字符):")
    print(f"   {redis_password}")
    print()
    
    # 3. TOKEN_SECRET_KEY
    token_key = generate_hex_key(32)  # 64字符
    print("3. TOKEN_SECRET_KEY (64字符):")
    print(f"   {token_key}")
    print()
    
    # 4. TCP_CRON_SECRET_KEY
    tcp_cron_key = generate_hex_key(16)  # 32字符
    print("4. TCP_CRON_SECRET_KEY (32字符):")
    print(f"   {tcp_cron_key}")
    print()
    
    # 5. TCP_AUTH_SECRET_KEY
    tcp_auth_key = generate_hex_key(16)  # 32字符
    print("5. TCP_AUTH_SECRET_KEY (32字符):")
    print(f"   {tcp_auth_key}")
    print()
    
    print("=" * 60)
    print("密码已生成！请妥善保管！")
    print("=" * 60)
    print()
    
    # 生成配置文件格式
    print("配置文件格式 (.env):")
    print("-" * 60)
    print(f"PG_PASSWORD={pg_password}")
    print(f"REDIS_PASSWORD={redis_password}")
    print(f"TOKEN_SECRET_KEY={token_key}")
    print(f"TCP_CRON_SECRET_KEY={tcp_cron_key}")
    print(f"TCP_AUTH_SECRET_KEY={tcp_auth_key}")
    print("-" * 60)
    print()
    
    # 生成YAML格式
    print("config.yaml 格式:")
    print("-" * 60)
    print(f"database:")
    print(f"  default:")
    print(f"    pass: \"{pg_password}\"")
    print()
    print(f"redis:")
    print(f"  default:")
    print(f"    pass: \"{redis_password}\"")
    print()
    print(f"token:")
    print(f"  secretKey: \"${{TOKEN_SECRET_KEY:{token_key}}}\"")
    print()
    print(f"tcp:")
    print(f"  client:")
    print(f"    cron:")
    print(f"      secretKey: \"${{TCP_CRON_SECRET_KEY:{tcp_cron_key}}}\"")
    print(f"    auth:")
    print(f"      secretKey: \"${{TCP_AUTH_SECRET_KEY:{tcp_auth_key}}}\"")
    print("-" * 60)

if __name__ == "__main__":
    main()

-- =============================================
-- 插入Bitget API配置
-- =============================================

INSERT INTO `hg_trading_api_config` (
    `tenant_id`,
    `user_id`,
    `api_name`,
    `platform`,
    `base_url`,
    `api_key`,
    `secret_key`,
    `passphrase`,
    `is_default`,
    `status`,
    `verify_status`,
    `verify_message`,
    `remark`,
    `created_at`,
    `updated_at`
) VALUES (
    0,                                                                          -- tenant_id
    1,                                                                          -- user_id (管理员)
    'Bitget-Main',                                                               -- api_name
    'bitget',                                                                   -- platform
    'https://api.bitget.com',                                                   -- base_url
    'bg_b7e03c26fabea7b74084e2526c462103',                                      -- api_key
    'e05dc0a32e174c2834f932ed907c5828e81707aba84829fe32840508de7b8c0e',          -- secret_key
    'lilei0680',                                                                -- passphrase
    1,                                                                          -- is_default
    1,                                                                          -- status (正常)
    1,                                                                          -- verify_status (验证成功)
    'Test OK: 21.68 USDT',                                                       -- verify_message
    'Bitget USDT-M Futures API',                                                 -- remark
    NOW(),                                                                      -- created_at
    NOW()                                                                       -- updated_at
);

SELECT '✅ Bitget API配置已插入！' as result;
SELECT * FROM `hg_trading_api_config` WHERE `platform` = 'bitget';


DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP RULE IF EXISTS audit_logs_no_delete ON audit_logs;
DROP RULE IF EXISTS audit_logs_no_update ON audit_logs;
DROP TABLE IF EXISTS audit_logs;

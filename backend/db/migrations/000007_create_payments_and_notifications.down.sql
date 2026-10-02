DROP TABLE IF EXISTS notification_deliveries;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS refunds;
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_payment;
DROP TABLE IF EXISTS payments;

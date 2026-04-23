-- Add transfer actions to inventory_logs constraint
ALTER TABLE inventory_logs DROP CONSTRAINT IF EXISTS inventory_logs_action_check;
ALTER TABLE inventory_logs ADD CONSTRAINT inventory_logs_action_check CHECK (action IN (
    'inbound', 'allocate', 'deallocate', 'pick', 'ship', 
    'replenish_out', 'replenish_in', 'adjust', 'return',
    'transfer_out', 'transfer_in'
));

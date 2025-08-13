-- Seed Exchange Rates for JeanPay Dashboard Testing
-- This script adds default exchange rates between NGN and GHS

-- Insert NGN to GHS exchange rate
INSERT INTO exchange_rates (
    from_currency,
    to_currency,
    rate,
    source,
    set_by,
    is_active,
    valid_from,
    created_at,
    updated_at
) VALUES (
    'NGN',
    'GHS',
    0.0053,
    'manual',
    'system',
    true,
    NOW(),
    NOW(),
    NOW()
) ON CONFLICT (from_currency, to_currency) DO UPDATE SET
    rate = EXCLUDED.rate,
    updated_at = NOW();

-- Insert GHS to NGN exchange rate
INSERT INTO exchange_rates (
    from_currency,
    to_currency,
    rate,
    source,
    set_by,
    is_active,
    valid_from,
    created_at,
    updated_at
) VALUES (
    'GHS',
    'NGN',
    188.68,
    'manual',
    'system',
    true,
    NOW(),
    NOW(),
    NOW()
) ON CONFLICT (from_currency, to_currency) DO UPDATE SET
    rate = EXCLUDED.rate,
    updated_at = NOW();

-- Verify the inserted rates
SELECT
    from_currency,
    to_currency,
    rate,
    is_active,
    created_at
FROM exchange_rates
WHERE (from_currency = 'NGN' AND to_currency = 'GHS')
   OR (from_currency = 'GHS' AND to_currency = 'NGN');

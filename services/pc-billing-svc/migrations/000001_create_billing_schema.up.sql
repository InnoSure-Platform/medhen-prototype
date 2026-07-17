CREATE TABLE billing_accounts (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL,
    account_type VARCHAR(20) NOT NULL DEFAULT 'DIRECT',
    suspense_balance NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    billing_account_id UUID NOT NULL REFERENCES billing_accounts(id),
    policy_id UUID NOT NULL,
    invoice_type VARCHAR(50) NOT NULL,
    total_amount NUMERIC(15,2) NOT NULL,
    amount_paid NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    coverage_start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    coverage_end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE invoice_line_items (
    id BIGSERIAL PRIMARY KEY,
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    description VARCHAR(255) NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    tax_amount NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    gl_account_code VARCHAR(50) NOT NULL
);

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    gateway_transaction_id VARCHAR(255) UNIQUE NOT NULL,
    method VARCHAR(50) NOT NULL,
    total_amount NUMERIC(15,2) NOT NULL,
    unallocated_amount NUMERIC(15,2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payment_allocations (
    id BIGSERIAL PRIMARY KEY,
    payment_id UUID NOT NULL REFERENCES payments(id),
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    amount NUMERIC(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ledger_transactions (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    reference_id UUID NOT NULL,
    reference_type VARCHAR(50) NOT NULL,
    posted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE journal_entries (
    id BIGSERIAL PRIMARY KEY,
    ledger_transaction_id UUID NOT NULL REFERENCES ledger_transactions(id),
    account_code VARCHAR(50) NOT NULL,
    debit_amount NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    credit_amount NUMERIC(15,2) NOT NULL DEFAULT 0.00
);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

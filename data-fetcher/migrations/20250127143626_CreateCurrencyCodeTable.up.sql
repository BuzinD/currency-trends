CREATE TABLE currencies
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique id
    code         VARCHAR(10)  NOT NULL,             -- Code (BTC, ETH)
    chain        VARCHAR(100) NOT NULL,             -- Network name USDT-TRC20, USDT-ERC20 etc.
    can_deposit  BOOLEAN      NOT NULL,             -- Can Deposit
    can_withdraw BOOLEAN      NOT NULL,             -- Can Withdraw
    UNIQUE (code, chain)
);

-- Idx for finding by code
CREATE INDEX idx_currency_code ON currencies (code);
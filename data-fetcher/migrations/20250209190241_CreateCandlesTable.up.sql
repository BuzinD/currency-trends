CREATE TABLE candles
(
    pair        VARCHAR(10) NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,
    open_price  BIGINT      NOT NULL,
    high_price  BIGINT      NOT NULL,
    low_price   BIGINT      NOT NULL,
    close_price BIGINT      NOT NULL,
    volume      BIGINT      NOT NULL,
    bar         VARCHAR(5)  NOT NULL,
    PRIMARY KEY (pair, timestamp, bar)
);

CREATE INDEX idx_pair_timestamp ON candles (pair, timestamp);
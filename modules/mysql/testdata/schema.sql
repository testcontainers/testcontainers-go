CREATE TABLE IF NOT EXISTS profile (
    id MEDIUMINT NOT NULL AUTO_INCREMENT,
    name VARCHAR(30) NOT NULL,
    PRIMARY KEY (id)
);

INSERT INTO profile (name) values ('profile 1');

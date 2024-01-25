CREATE TABLE servers_mc (
    id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
    srvid varchar(8) NOT NULL,
    srv_name varchar(255) NOT NULL,
    plan varchar(8) NOT NULL,
    owner_id int NOT NULL,
    creation_date datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expire_date datetime NOT NULL,
    version varchar(32) NOT NULL,
    core varchar(64) NOT NULL,
    ram_min int NOT NULL DEFAULT 0,
    ram_max int NOT NULL DEFAULT 0,
    cpus int NOT NULL DEFAULT 0,
    disk int NOT NULL DEFAULT 0,
    description varchar(255) NOT NULL DEFAULT '',
    icon varchar(16) NOT NULL DEFAULT 'mc_default.png',

    FOREIGN KEY (owner_id) REFERENCES users(uid)
)
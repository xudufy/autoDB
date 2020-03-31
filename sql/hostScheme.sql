drop database if exists autodb;
create database autodb;

use autodb;

-- drop table if exists users;
-- drop table if exists projects;
-- drop table if exists project_developer;
-- drop table if exists tables;
-- drop table if exists apis;
-- drop table if exists projects;

create table users (
	uid int primary key auto_increment,
    email varchar(100) not null unique,
    pw char(64) not null, -- sha256 result
    username varchar(100) not null
);

create table projects (
	pid int auto_increment primary key,
    pname varchar(100) unique not null,
    pw char(64) not null,
    create_time datetime default CURRENT_TIMESTAMP,
    check ( pname <> 'autodb' )
);

create table project_developer (
	uid int,
    pid int,
	privilege enum('owner', 'developer') not null,
    primary key (uid, pid),
    foreign key (uid) references users(uid),
    foreign key (pid) references projects(pid) on delete cascade
);

DELIMITER $$

create trigger num_owner_check_delete before delete 
	on project_developer
for each row
begin
	declare owner_num int;
	if (old.privilege='owner') then
		set owner_num =  (select count(*) from project_developer where pid = old.pid and privilege='owner') ;
		if owner_num = 1 THEN
			SIGNAL SQLSTATE '45000' set message_text='must have one owner.';
		end if;
    end if;
end$$

create trigger num_owner_check_update before update 
	on project_developer
for each row
begin
	declare owner_num int;
	if (new.privilege<>'owner' and old.privilege='owner') then
		set owner_num = (select count(*) from project_developer where pid = old.pid and privilege='owner');
		if owner_num = 1 THEN
			SIGNAL SQLSTATE '45000' set message_text='must have one owner.';
		end if;
    end if;
end$$

DELIMITER ;

create table tables(
	tid int auto_increment primary key,
    pid int not null,
    name varchar(64) not null,
    unique (pid, name),
    foreign key (pid) references projects(pid) on delete cascade,
    index (pid, name)
);

create table apis (
	aid char(64) unique, -- sha256 result
	tid int not null,
    name varchar(64) not null,
    type enum('public', 'user-domain', 'developer-domain') not null,
	tmpl varchar(8192),
    primary key (tid, name),
    foreign key (tid) references tables(tid) on delete cascade,
    index (aid)
)

create table api_parameters (
    aid char(64) unique,
    name varchar(64),
    input_type enum('string', 'time', 'int', 'double'),
    primary key (aid, name),
    foreign key (aid) references apis(aid) on delete cascade
)

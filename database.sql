-- auto-generated definition
create table hospital_status
(
    id                        serial not null,
    district                  varchar(500),
    institution               varchar,
    covid_beds                json,
    oxygen_supported_beds     json,
    non_oxygen_supported_beds json,
    icu_beds                  json,
    ventilator                json,
    last_updated              varchar,
    contact_number            varchar,
    remarks                   text,
    updated_at                timestamp default CURRENT_TIMESTAMP,
    created_at                timestamp default CURRENT_TIMESTAMP,
    constraint tn_status_district_institution_key
        unique (district, institution)
);

alter table hospital_status
    owner to backend;

create unique index tn_status_institution_uindex
    on hospital_status (institution);

-- auto-generated definition
create table push_logs
(
    id     bigserial not null,
    title  text,
    body   text,
    page   text,
    status varchar
);

alter table push_logs
    owner to backend;

-- auto-generated definition
create table push_subscription
(
    id            serial not null,
    email         varchar,
    subscriptions integer[],
    created_at    timestamp default CURRENT_TIMESTAMP,
    updated_at    timestamp default CURRENT_TIMESTAMP,
    token         text
);

alter table push_subscription
    owner to backend;


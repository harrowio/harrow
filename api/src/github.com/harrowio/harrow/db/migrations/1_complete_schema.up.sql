-- +migrate Up
-- +migrate StatementBegin


--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.4
-- Dumped by pg_dump version 9.6.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: btree_gist; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;


--
-- Name: EXTENSION btree_gist; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION btree_gist IS 'support for indexing common datatypes in GiST';


--
-- Name: hstore; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS hstore WITH SCHEMA public;


--
-- Name: EXTENSION hstore; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION hstore IS 'data type for storing sets of (key, value) pairs';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

--
-- Name: harrow_context_user; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE harrow_context_user AS (
	uuid uuid,
	name text,
	email text,
	url_host text
);



--
-- Name: membership_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE membership_type AS ENUM (
    'guest',
    'member',
    'manager',
    'owner'
);



--
-- Name: oauth_provider; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE oauth_provider AS ENUM (
    'github',
    'bitbucket'
);



--
-- Name: operation_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE operation_type AS ENUM (
    'git.check.access',
    'git.enumeration.branches',
    'git.enumeration.commits',
    'job.scheduled',
    'job.webhooked',
    'job.retried',
    'notifier.invoke'
);



--
-- Name: repository_credential_status; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE repository_credential_status AS ENUM (
    'pending',
    'present'
);



--
-- Name: repository_credential_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE repository_credential_type AS ENUM (
    'ssh',
    'basic'
);



--
-- Name: schedule_disabled_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE schedule_disabled_type AS ENUM (
    'internal_error',
    'job_archived',
    'ran_once'
);



--
-- Name: secret_status; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE secret_status AS ENUM (
    'pending',
    'present'
);



--
-- Name: secret_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE secret_type AS ENUM (
    'ssh',
    'env'
);



--
-- Name: target_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE target_type AS ENUM (
    's3compat',
    'ssh',
    'git',
    'ftp',
    'email',
    'webhook'
);



--
-- Name: task_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE task_type AS ENUM (
    'script',
    'test',
    'build',
    'deployment'
);



--
-- Name: workspace_base_image_type; Type: TYPE; Schema: public; Owner: production
--

CREATE TYPE workspace_base_image_type AS ENUM (
    'container'
);



--
-- Name: broadcast_change(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION broadcast_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
        message text := json_build_object(
            'table', TG_TABLE_NAME,
            'new', row_to_json(NEW),
            'old', row_to_json(OLD),
            'contextUser', context_user()
        );
        message_length int := octet_length(message);
        message_maximum int := 8000;
BEGIN
IF TG_TABLE_NAME = 'operations' THEN
    message := json_build_object(
        'table', TG_TABLE_NAME,
        'new', json_build_object(
            'uuid', NEW.uuid,
            'created_at', NEW.created_at,
            'started_at', NEW.started_at,
            'timed_out_at', NEW.timed_out_at,
            'exit_status', NEW.exit_status
        ),
        'old', json_build_object(
            'uuid', OLD.uuid,
            'created_at', OLD.created_at,
            'started_at', OLD.started_at,
            'timed_out_at', OLD.timed_out_at,
            'exit_status', OLD.exit_status
        ),
        'contextUser', context_user()
    );
END IF;

IF message_length >= message_maximum THEN
    message := json_build_object(
        'table', TG_TABLE_NAME,
        'new', json_build_object(
            'uuid', NEW.uuid
        ),
        'old', json_build_object(
            'uuid', OLD.uuid
        ),
        'contextUser', context_user()
    );
END IF;

PERFORM pg_notify('broadcast/change', message);
RETURN NEW;
END;
$$;



--
-- Name: broadcast_create(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION broadcast_create() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
        message text := json_build_object(
            'table', TG_TABLE_NAME,
            'new', row_to_json(NEW),
            'contextUser', context_user()
        );
        message_length int := octet_length(message);
        message_maximum int := 8000;
BEGIN
IF message_length >= message_maximum THEN
    IF TG_TABLE_NAME = 'activities' THEN
        message := json_build_object(
            'table', TG_TABLE_NAME,
            'new', json_build_object(
                'id', NEW.id
            ),
            'contextUser', context_user()
        );
    ELSE
        message := json_build_object(
            'table', TG_TABLE_NAME,
            'new', json_build_object(
                'uuid', NEW.uuid
            ),
            'contextUser', context_user()
        );
    END IF;
END IF;

PERFORM pg_notify('broadcast/create', message);
RETURN NEW;
END;
$$;



--
-- Name: broadcast_create_uuid(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION broadcast_create_uuid() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
      BEGIN
        PERFORM pg_notify(CAST('broadcast/create' AS text), CAST('{"table": "' || TG_TABLE_NAME || '", "new": ' || json_build_object('uuid', NEW.uuid) || ', "contextUser": ' || context_user() || '}' AS text));
        RETURN NEW;
      END;
    $$;



--
-- Name: broadcast_session_change(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION broadcast_session_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    message text;
BEGIN
    message := json_build_object(
        'table', TG_TABLE_NAME,
        'new', json_build_object(
            'uuid', NEW.uuid
        ),
        'old', json_build_object(
            'uuid', OLD.uuid
        ),
        'contextUser', context_user()
    );

    IF OLD.validated_at = NEW.validated_at
    THEN
        PERFORM pg_notify('broadcast/change', message);
    END IF;
RETURN NEW;
END;
$$;



--
-- Name: cascade_environment_archival(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION cascade_environment_archival() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE jobs    SET archived_at = new.archived_at WHERE environment_uuid = new.uuid AND archived_at IS NULL;
  UPDATE secrets SET archived_at = new.archived_at WHERE environment_uuid = new.uuid AND archived_at IS NULL;
  RETURN new;
END;
$$;



--
-- Name: cascade_job_archival(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION cascade_job_archival() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE schedules SET disabled = 'job_archived', disabled_because = NULL
    WHERE job_uuid = new.uuid
    AND   disabled IS NULL;
  UPDATE operations SET archived_at = new.archived_at
    WHERE job_uuid = new.uuid
    AND   archived_at IS NULL;
  UPDATE webhooks SET archived_at = new.archived_at
    WHERE job_uuid = new.uuid
    AND   archived_at IS NULL;
  UPDATE git_triggers SET archived_at = new.archived_at
    WHERE job_uuid = new.uuid
    AND   archived_at IS NULL;
  UPDATE notification_rules SET archived_at = new.archived_at
   WHERE job_uuid = new.uuid
    AND  archived_at IS NULL;
  UPDATE job_notifiers SET archived_at = new.archived_at
   WHERE (
      job_uuid = new.uuid
   )
    AND  archived_at IS NULL;
  RETURN new;
END;
$$;



--
-- Name: cascade_organization_archival(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION cascade_organization_archival() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE projects SET archived_at = new.archived_at WHERE organization_uuid = new.uuid AND archived_at IS NULL;

  RETURN new;
END;
$$;



--
-- Name: cascade_project_archival(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION cascade_project_archival() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE environments        SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  UPDATE invitations         SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  UPDATE jobs                SET archived_at = new.archived_at WHERE task_uuid IN (SELECT uuid FROM tasks WHERE project_uuid = new.uuid) AND archived_at IS NULL;
  UPDATE project_memberships SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  UPDATE repositories        SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  UPDATE subscriptions       SET archived_at = new.archived_at WHERE watchable_uuid = new.uuid AND watchable_type = 'project' AND archived_at IS NULL;
  UPDATE targets             SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  UPDATE tasks               SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  UPDATE webhooks            SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  UPDATE git_triggers        SET archived_at = new.archived_at WHERE project_uuid = new.uuid AND archived_at IS NULL;
  RETURN new;
END;
$$;



--
-- Name: cascade_repository_archival(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION cascade_repository_archival() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE operations SET archived_at = new.archived_at WHERE repository_uuid = new.uuid AND archived_at IS NULL;
  UPDATE repository_credentials SET archived_at = new.archived_at WHERE repository_uuid = new.uuid AND archived_at IS NULL;
  RETURN new;
END;
$$;



--
-- Name: cascade_task_archival(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION cascade_task_archival() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE jobs SET archived_at = new.archived_at WHERE task_uuid = new.uuid AND archived_at IS NULL;

  RETURN new;
END;
$$;



--
-- Name: cascade_webhook_archival(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION cascade_webhook_archival() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  UPDATE deliveries SET archived_at = new.archived_at WHERE webhook_uuid = new.uuid AND archived_at IS NULL;
  RETURN new;
END;
$$;



--
-- Name: context_user(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION context_user() RETURNS SETOF text
    LANGUAGE plpgsql
    AS $$
			BEGIN
				BEGIN
					-- Inelegantly count the number of users, otherwise return 'null'
					IF (SELECT COUNT(*) FROM users WHERE uuid = current_setting('harrow.context_user_uuid')::uuid LIMIT 1) != 1 THEN
						RETURN QUERY SELECT 'null'::text;
					ELSE
						RETURN QUERY SELECT row_to_json(
                                                    row(uuid, name, email, url_host)::harrow_context_user
                                                )::text
                                                FROM users
                                                WHERE uuid = current_setting('harrow.context_user_uuid')::uuid LIMIT 1;
					END IF;
				EXCEPTION
					WHEN SQLSTATE '42704' THEN RETURN QUERY SELECT 'null'::text;
						-- SQLSTATE 'undefined_object' current_setting() not found
					WHEN SQLSTATE '42883' THEN RETURN QUERY SELECT 'null'::text;
						-- SQLSTATE 'undefined_function' current_setting() does not contain a well formed UUID
					WHEN SQLSTATE '22P02' THEN RETURN QUERY SELECT 'null'::text;
						-- SQLSTATE 'invalid_text_representation' current_setting() does not contain a well formed UUID
				END;
			END
		$$;



--
-- Name: is_active_user_job(timestamp without time zone, text); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION is_active_user_job(archived_at timestamp without time zone, name text) RETURNS boolean
    LANGUAGE plpgsql IMMUTABLE
    AS $$ begin return archived_at is null and name::text !~~ 'urn:harrow%'; end; $$;



--
-- Name: is_active_user_job(timestamp with time zone, character varying); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION is_active_user_job(archived_at timestamp with time zone, name character varying) RETURNS boolean
    LANGUAGE plpgsql IMMUTABLE
    AS $$ begin return archived_at is null and name::text !~~ 'urn:harrow%'; end; $$;



--
-- Name: is_job_operation(operation_type); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION is_job_operation(t operation_type) RETURNS boolean
    LANGUAGE plpgsql IMMUTABLE
    AS $$ begin return t::text like 'job.%'; end; $$;



--
-- Name: is_valid_timezone_name(character varying); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION is_valid_timezone_name(character varying) RETURNS boolean
    LANGUAGE plpgsql
    AS $_$
DECLARE
    ignore integer := 0;
BEGIN
SELECT 1 INTO ignore FROM harrow_timezones WHERE name = $1;
IF NOT FOUND THEN
    RETURN FALSE;
ELSE
    RETURN TRUE;
END IF;
END;
$_$;



--
-- Name: json_object_delete_key(json, text); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION json_object_delete_key(object json, key_to_delete text) RETURNS json
    LANGUAGE sql IMMUTABLE STRICT
    AS $$
SELECT json_object_agg(key, value) FROM (
       SELECT * FROM json_each(object)
       WHERE  key <> key_to_delete
) AS returned_object;
$$;



--
-- Name: json_object_set_key(json, text, anyelement); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION json_object_set_key(object json, key_to_set text, value_to_set anyelement) RETURNS json
    LANGUAGE sql IMMUTABLE STRICT
    AS $$
SELECT json_object_agg(key, value) FROM (
       SELECT * FROM json_each(object)
       WHERE  key <> key_to_set
       UNION ALL
       SELECT key_to_set, to_json(value_to_set)
) AS returned_object;
$$;



--
-- Name: kpi_stats(timestamp without time zone, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION kpi_stats(from_date timestamp without time zone, to_date timestamp without time zone) RETURNS TABLE(organizations integer, deleted_organizations integer, organizations_created_this_week integer, active_organizations_1 integer, active_organizations_5 integer, active_organizations_15 integer, new_organizations_last_week_1 integer, new_organizations_last_week_5 integer, new_organizations_last_week_15 integer, organizations_silver integer, organizations_gold integer, organizations_platinum integer, organizations_converted_to_paying integer)
    LANGUAGE plpgsql
    AS $_$
DECLARE
 calendar_week tstzrange;
 previous_calendar_week tstzrange;
BEGIN
 calendar_week := ('[' || $1 || ',' || $2 || ']')::tstzrange;
 previous_calendar_week := ('[' || ($1 - interval '7 days') || ',' || ($2 - interval '7 days') || ']');

 organizations := (SELECT COUNT(*) FROM organizations WHERE archived_at IS NULL AND created_at <= to_date);
 deleted_organizations := (SELECT COUNT(*) FROM organizations WHERE archived_at IS NOT NULL AND created_at <= to_date);
 organizations_created_this_week := (SELECT COUNT(*) FROM organizations WHERE created_at <@ calendar_week);
 active_organizations_1 := (SELECT COUNT(*)
 FROM
   (SELECT  oa.organization_uuid, COUNT(a.*)
      FROM  activities AS a,
            organization_activities AS oa
      WHERE a.occurred_on <@ calendar_week
        AND a.id = oa.id
        AND oa.organization_uuid IS NOT NULL
   GROUP BY 1
     HAVING COUNT(a.*) > 1
   ORDER BY 2 DESC
   ) data
 );
 active_organizations_5 := (SELECT COUNT(*)
 FROM
   (SELECT  oa.organization_uuid, COUNT(a.*)
      FROM  activities AS a,
            organization_activities AS oa
      WHERE a.occurred_on <@ calendar_week
        AND a.id = oa.id
        AND oa.organization_uuid IS NOT NULL
   GROUP BY 1
     HAVING COUNT(a.*) > 5
   ORDER BY 2 DESC
   ) data
 );
 active_organizations_15 := (SELECT COUNT(*)
 FROM
   (SELECT  oa.organization_uuid, COUNT(a.*)
      FROM  activities AS a,
            organization_activities AS oa
      WHERE a.occurred_on <@ calendar_week
        AND a.id = oa.id
        AND oa.organization_uuid IS NOT NULL
   GROUP BY 1
     HAVING COUNT(a.*) > 15
   ORDER BY 2 DESC
   ) data
 );

 new_organizations_last_week_1 := (SELECT COUNT(*) FROM(
   SELECT oa.organization_uuid,
          COUNT(a.*)
     FROM activities AS a,
          organization_activities AS oa,
          organizations o
    WHERE a.occurred_on <@ calendar_week
      AND a.id = oa.id
      AND oa.organization_uuid IS NOT NULL
      AND oa.organization_uuid = o.uuid
      AND o.created_at <@ previous_calendar_week
 GROUP BY 1
  HAVING count(a.*) > 1
 ORDER BY 2 DESC
 ) data);

 new_organizations_last_week_5 := (SELECT COUNT(*) FROM(
   SELECT oa.organization_uuid,
          COUNT(a.*)
     FROM activities AS a,
          organization_activities AS oa,
          organizations o
    WHERE a.occurred_on <@ calendar_week
      AND a.id = oa.id
      AND oa.organization_uuid IS NOT NULL
      AND oa.organization_uuid = o.uuid
      AND o.created_at <@ previous_calendar_week
 GROUP BY 1
  HAVING count(a.*) > 5
 ORDER BY 2 DESC
 ) data);

 new_organizations_last_week_15 := (SELECT COUNT(*) FROM(
   SELECT oa.organization_uuid,
          COUNT(a.*)
     FROM activities AS a,
          organization_activities AS oa,
          organizations o
    WHERE a.occurred_on <@ calendar_week
      AND a.id = oa.id
      AND oa.organization_uuid IS NOT NULL
      AND oa.organization_uuid = o.uuid
      AND o.created_at <@ previous_calendar_week
 GROUP BY 1
  HAVING count(a.*) > 15
 ORDER BY 2 DESC
 ) data);

 organizations_silver := (SELECT COUNT(*) FROM (
     SELECT o.uuid,
            o.name,
            (SELECT lower((data->'PlanName')::text)
               FROM billing_events
              WHERE organization_uuid = o.uuid
                AND event_name = 'plan-selected'
                AND occurred_on < to_date
           ORDER BY occurred_on DESC
              LIMIT 1
             ) AS plan_name FROM organizations o
 ) data WHERE plan_name = '"silver"');

 organizations_gold := (SELECT COUNT(*) FROM (
     SELECT o.uuid,
            o.name,
            (SELECT lower((data->'PlanName')::text)
               FROM billing_events
              WHERE organization_uuid = o.uuid
                AND event_name = 'plan-selected'
                AND occurred_on < to_date
           ORDER BY occurred_on DESC
              LIMIT 1
             ) AS plan_name FROM organizations o
 ) data WHERE plan_name = '"gold"');

 organizations_platinum := (SELECT COUNT(*) FROM (
     SELECT o.uuid,
            o.name,
            (SELECT lower((data->'PlanName')::text)
               FROM billing_events
              WHERE organization_uuid = o.uuid
                AND event_name = 'plan-selected'
                AND occurred_on < to_date
           ORDER BY occurred_on DESC
              LIMIT 1
             ) AS plan_name FROM organizations o
 ) data WHERE plan_name = '"platinum"');

 organizations_converted_to_paying := (SELECT COUNT(*) FROM (
   SELECT uuid as organization_uuid
     FROM organizations AS o
    WHERE created_at >= (from_date - interval '21 days')
      AND created_at <= (from_date - interval '14 days')
      AND (SELECT COUNT(*)
             FROM billing_events be
            WHERE be.organization_uuid = o.uuid
              AND be.event_name = 'plan-selected'
              AND lower( (be.data->'PlanName')::text ) != '"free"'
              AND be.occurred_on < to_date
         GROUP BY be.occurred_on
         ORDER BY be.occurred_on DESC
            LIMIT 1
              ) = 1
 ) data );

 RETURN QUERY SELECT organizations, deleted_organizations, organizations_created_this_week, active_organizations_1, active_organizations_5, active_organizations_15, new_organizations_last_week_1, new_organizations_last_week_5, new_organizations_last_week_15, organizations_silver, organizations_gold, organizations_platinum, organizations_converted_to_paying;
END;
$_$;



--
-- Name: notify_new_operation(); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION notify_new_operation() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
      BEGIN
        PERFORM pg_notify('new-operation', NULL);
        RETURN NEW;
      END;
    $$;



--
-- Name: put_organization_on_free_plan(uuid, uuid); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION put_organization_on_free_plan(user_uuid uuid, organization_uuid uuid) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
  subscription_id text;
  payload text;
  event_uuid uuid;
BEGIN
  subscription_id := uuid_generate_v4()::text || '-free-' || now()::text;
  payload := '{"UserUuid":"' || user_uuid || '","PlanUuid":"b99a21cc-b108-466e-aa4d-bde10ebbe1f3","PlanName":"Free","SubscriptionId":"' || subscription_id || '","PrivateCodeAvailable":true,"PricePerMonth":"0.00 USD","UsersIncluded":1,"ProjectsIncluded":1,"PricePerAdditionalUser":"0.00 USD","NumberOfConcurrentJobs":1}';
  INSERT INTO billing_events (organization_uuid, event_name, data) values (organization_uuid, 'plan-selected', payload::json) RETURNING uuid INTO event_uuid;  
  RETURN event_uuid;
END;
$$;



--
-- Name: random_string(integer); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION random_string(length integer) RETURNS text
    LANGUAGE plpgsql
    AS $$
declare
  chars text[] := '{0,1,2,3,4,5,6,7,8,9,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z}';
  result text := '';
  i integer := 0;
begin
  if length < 0 then
    raise exception 'Given length cannot be less than 0';
  end if;
  for i in 1..length loop
    result := result || chars[1+random()*(array_length(chars, 1)-1)];
  end loop;
  return result;
end;
$$;



--
-- Name: round_time(timestamp with time zone); Type: FUNCTION; Schema: public; Owner: production
--

CREATE FUNCTION round_time(timestamp with time zone) RETURNS timestamp with time zone
    LANGUAGE sql
    AS $_$
	  SELECT date_trunc('hour', $1) + INTERVAL '5 min' * ROUND(date_part('minute', $1) / 5.0)
		$_$;



SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: activities; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE activities (
    id bigint NOT NULL,
    name text NOT NULL,
    occurred_on timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    context_user_uuid uuid,
    payload json,
    extra json
);



--
-- Name: TABLE activities; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON TABLE activities IS 'History of noteworthy events triggered by users or programs';


--
-- Name: COLUMN activities.name; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON COLUMN activities.name IS 'Identifier for the type of activity, e.g. job.added, job.edited';


--
-- Name: COLUMN activities.occurred_on; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON COLUMN activities.occurred_on IS 'Time at which the activity took place';


--
-- Name: COLUMN activities.created_at; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON COLUMN activities.created_at IS 'Time at which the activity has been persisted';


--
-- Name: COLUMN activities.context_user_uuid; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON COLUMN activities.context_user_uuid IS 'Id of the user who caused the activity.  Not set if the activity was caused by a program';


--
-- Name: COLUMN activities.payload; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON COLUMN activities.payload IS 'Context dependent data about the activity';


--
-- Name: COLUMN activities.extra; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON COLUMN activities.extra IS 'Additional data added to the activity during processing';


--
-- Name: activities_archive; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE activities_archive (
    id bigint,
    name text,
    occurred_on timestamp with time zone,
    created_at timestamp with time zone,
    context_user_uuid uuid,
    payload json,
    extra json
);



--
-- Name: activities_id_seq; Type: SEQUENCE; Schema: public; Owner: production
--

CREATE SEQUENCE activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: production
--

ALTER SEQUENCE activities_id_seq OWNED BY activities.id;


--
-- Name: billing_events; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE billing_events (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    organization_uuid uuid NOT NULL,
    event_name text NOT NULL,
    occurred_on timestamp with time zone DEFAULT now(),
    data json NOT NULL
);



--
-- Name: broken_activities_i_deleted; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE broken_activities_i_deleted (
    id bigint,
    name text,
    occurred_on timestamp with time zone,
    created_at timestamp with time zone,
    context_user_uuid uuid,
    payload json,
    extra json
);



--
-- Name: deliveries; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE deliveries (
    uuid uuid NOT NULL,
    webhook_uuid uuid NOT NULL,
    delivered_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    request text NOT NULL,
    schedule_uuid uuid,
    archived_at timestamp with time zone
);



--
-- Name: disabled_recurring_scheduled; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE disabled_recurring_scheduled (
    uuid uuid
);



--
-- Name: duplicate_jobs; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE duplicate_jobs (
    task_uuid uuid,
    environment_uuid uuid,
    count bigint,
    uuids uuid[]
);



--
-- Name: email_notifiers; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE email_notifiers (
    uuid uuid NOT NULL,
    recipient text NOT NULL,
    archived_at timestamp with time zone,
    url_host text DEFAULT 'www.vm.harrow.io'::text NOT NULL,
    project_uuid uuid
);



--
-- Name: environments; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE environments (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    project_uuid uuid NOT NULL,
    variables hstore,
    archived_at timestamp with time zone,
    is_default boolean DEFAULT false NOT NULL
);



--
-- Name: git_triggers; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE git_triggers (
    uuid uuid NOT NULL,
    name text NOT NULL,
    project_uuid uuid NOT NULL,
    job_uuid uuid NOT NULL,
    repository_uuid uuid,
    change_type text NOT NULL,
    match_ref text NOT NULL,
    archived_at timestamp with time zone,
    creator_uuid uuid NOT NULL
);



--
-- Name: harrow_timezones; Type: MATERIALIZED VIEW; Schema: public; Owner: production
--

CREATE MATERIALIZED VIEW harrow_timezones AS
 SELECT pg_timezone_names.name,
    pg_timezone_names.abbrev,
    pg_timezone_names.utc_offset,
    pg_timezone_names.is_dst
   FROM pg_timezone_names
  WITH NO DATA;

REFRESH MATERIALIZED VIEW harrow_timezones;


--
-- Name: hung_ops_2017_01_19; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE hung_ops_2017_01_19 (
    uuid uuid,
    type operation_type,
    workspace_base_image_uuid uuid,
    repository_uuid uuid,
    time_limit integer,
    exit_status integer,
    created_at timestamp with time zone,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    timed_out_at timestamp with time zone,
    fatal_error text,
    job_uuid uuid,
    failed_at timestamp with time zone,
    archived_at timestamp with time zone,
    parameters json,
    repository_refs json,
    git_logs json,
    status_logs json,
    notifier_uuid uuid,
    notifier_type text,
    canceled_at timestamp without time zone
);



--
-- Name: hung_ops_2017_01_19__0002; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE hung_ops_2017_01_19__0002 (
    uuid uuid,
    type operation_type,
    workspace_base_image_uuid uuid,
    repository_uuid uuid,
    time_limit integer,
    exit_status integer,
    created_at timestamp with time zone,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    timed_out_at timestamp with time zone,
    fatal_error text,
    job_uuid uuid,
    failed_at timestamp with time zone,
    archived_at timestamp with time zone,
    parameters json,
    repository_refs json,
    git_logs json,
    status_logs json,
    notifier_uuid uuid,
    notifier_type text,
    canceled_at timestamp without time zone
);



--
-- Name: hung_ops_2017_01_25; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE hung_ops_2017_01_25 (
    uuid uuid,
    type operation_type,
    workspace_base_image_uuid uuid,
    repository_uuid uuid,
    time_limit integer,
    exit_status integer,
    created_at timestamp with time zone,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    timed_out_at timestamp with time zone,
    fatal_error text,
    job_uuid uuid,
    failed_at timestamp with time zone,
    archived_at timestamp with time zone,
    parameters json,
    repository_refs json,
    git_logs json,
    status_logs json,
    notifier_uuid uuid,
    notifier_type text,
    canceled_at timestamp without time zone
);



--
-- Name: hung_ops_2017_01_25_02; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE hung_ops_2017_01_25_02 (
    uuid uuid,
    type operation_type,
    workspace_base_image_uuid uuid,
    repository_uuid uuid,
    time_limit integer,
    exit_status integer,
    created_at timestamp with time zone,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    timed_out_at timestamp with time zone,
    fatal_error text,
    job_uuid uuid,
    failed_at timestamp with time zone,
    archived_at timestamp with time zone,
    parameters json,
    repository_refs json,
    git_logs json,
    status_logs json,
    notifier_uuid uuid,
    notifier_type text,
    canceled_at timestamp without time zone
);



--
-- Name: invitations; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE invitations (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    email character varying(254) NOT NULL,
    organization_uuid uuid NOT NULL,
    project_uuid uuid NOT NULL,
    membership_type membership_type DEFAULT 'member'::membership_type NOT NULL,
    creator_uuid uuid NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    sent_at timestamp with time zone,
    accepted_at timestamp with time zone,
    invitee_uuid uuid NOT NULL,
    recipient_name text NOT NULL,
    message text NOT NULL,
    refused_at timestamp with time zone,
    archived_at timestamp with time zone
);



--
-- Name: COLUMN invitations.invitee_uuid; Type: COMMENT; Schema: public; Owner: production
--

COMMENT ON COLUMN invitations.invitee_uuid IS '
used as user id for new users signing up due to an invitation
';


--
-- Name: job_notifiers; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE job_notifiers (
    uuid uuid NOT NULL,
    webhook_url text NOT NULL,
    job_uuid uuid NOT NULL,
    archived_at timestamp with time zone,
    project_uuid uuid NOT NULL
);



--
-- Name: jobs; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE jobs (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    task_uuid uuid NOT NULL,
    environment_uuid uuid NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    archived_at timestamp with time zone,
    description text
);



--
-- Name: projects; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE projects (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    organization_uuid uuid NOT NULL,
    public boolean DEFAULT false NOT NULL,
    archived_at timestamp with time zone
);



--
-- Name: tasks; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE tasks (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    project_uuid uuid NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    body text NOT NULL,
    type task_type DEFAULT 'script'::task_type NOT NULL,
    archived_at timestamp with time zone
);



--
-- Name: jobs_projects; Type: VIEW; Schema: public; Owner: production
--

CREATE VIEW jobs_projects AS
 SELECT DISTINCT j.uuid,
        CASE
            WHEN ((j.name)::text ~~ 'urn:%'::text) THEN (j.name)::text
            ELSE (((e.name)::text || ' - '::text) || (t.name)::text)
        END AS name,
    j.task_uuid,
    j.environment_uuid,
    j.created_at,
    j.archived_at,
    p.uuid AS project_uuid,
    p.name AS project_name
   FROM (((jobs j
     JOIN environments e ON ((j.environment_uuid = e.uuid)))
     JOIN tasks t ON ((j.task_uuid = t.uuid)))
     JOIN projects p ON ((t.project_uuid = p.uuid)));



--
-- Name: notification_rules; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE notification_rules (
    uuid uuid NOT NULL,
    project_uuid uuid NOT NULL,
    notifier_uuid uuid NOT NULL,
    notifier_type text NOT NULL,
    match_activity text NOT NULL,
    creator_uuid uuid NOT NULL,
    archived_at timestamp with time zone,
    job_uuid uuid
);



--
-- Name: oauth_tokens; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE oauth_tokens (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    provider oauth_provider NOT NULL,
    access_token character varying(100) NOT NULL,
    token_type character varying(100) NOT NULL,
    scopes character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    user_uuid uuid NOT NULL
);



--
-- Name: organization_memberships; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE organization_memberships (
    organization_uuid uuid NOT NULL,
    user_uuid uuid NOT NULL,
    type membership_type DEFAULT 'member'::membership_type NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);



--
-- Name: organizations; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE organizations (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    github_login character varying(100) DEFAULT ''::character varying,
    public boolean DEFAULT false NOT NULL,
    archived_at timestamp with time zone
);



--
-- Name: project_memberships; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE project_memberships (
    uuid uuid NOT NULL,
    project_uuid uuid NOT NULL,
    user_uuid uuid NOT NULL,
    membership_type membership_type,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    archived_at timestamp with time zone
);



--
-- Name: old_organizations; Type: VIEW; Schema: public; Owner: production
--

CREATE VIEW old_organizations AS
 SELECT o.uuid,
    o.created_at,
    o.name,
    ( SELECT count(*) AS count
           FROM projects
          WHERE ((projects.organization_uuid = o.uuid) AND (projects.archived_at IS NULL))) AS projects,
    ( SELECT count(*) AS count
           FROM ( SELECT organization_memberships.user_uuid
                   FROM organization_memberships
                  WHERE ((organization_memberships.organization_uuid = o.uuid) AND (o.archived_at IS NULL))
                UNION
                 SELECT pm.user_uuid
                   FROM project_memberships pm,
                    projects p
                  WHERE ((p.organization_uuid = o.uuid) AND (pm.archived_at IS NULL) AND (pm.project_uuid = p.uuid))) membership_counts) AS members
   FROM organizations o
  WHERE ((( SELECT count(*) AS count
           FROM billing_events
          WHERE ((billing_events.organization_uuid = o.uuid) AND (billing_events.event_name = 'plan-selected'::text))) = 0) AND (o.archived_at IS NULL));



--
-- Name: operations; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE operations (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    type operation_type NOT NULL,
    workspace_base_image_uuid uuid NOT NULL,
    repository_uuid uuid,
    time_limit integer DEFAULT 300 NOT NULL,
    exit_status integer DEFAULT 256 NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    timed_out_at timestamp with time zone,
    fatal_error text,
    job_uuid uuid,
    failed_at timestamp with time zone,
    archived_at timestamp with time zone,
    parameters json,
    repository_refs json,
    git_logs json,
    status_logs json,
    notifier_uuid uuid,
    notifier_type text,
    canceled_at timestamp without time zone,
    CONSTRAINT operations_check CHECK (((repository_uuid IS NOT NULL) OR (notifier_uuid IS NOT NULL) OR (job_uuid IS NOT NULL)))
);



--
-- Name: operation_90th_percentile_start_wait_time; Type: VIEW; Schema: public; Owner: production
--

CREATE VIEW operation_90th_percentile_start_wait_time AS
 SELECT operations.created_at,
    operations.started_at,
    (operations.started_at - operations.created_at) AS wait_time
   FROM operations
  WHERE (operations.job_uuid IS NOT NULL);



--
-- Name: operation_statistics; Type: VIEW; Schema: public; Owner: production
--

CREATE VIEW operation_statistics AS
 SELECT operations.created_at,
    operations.started_at,
    (operations.started_at - operations.created_at) AS wait_time,
    (operations.finished_at - operations.started_at) AS run_time,
    (operations.timed_out_at - operations.started_at) AS time_out_time,
    (operations.failed_at - operations.started_at) AS failed_time,
    (operations.job_uuid IS NOT NULL) AS is_user,
    (operations.repository_uuid IS NOT NULL) AS is_repo
   FROM operations;



--
-- Name: organization_activities; Type: MATERIALIZED VIEW; Schema: public; Owner: production
--

CREATE MATERIALIZED VIEW organization_activities AS
 SELECT o.uuid AS organization_uuid,
    activities.id
   FROM activities,
    organizations o
  WHERE (((activities.extra -> 'organizationUuid'::text))::text = (('"'::text || o.uuid) || '"'::text))
UNION
 SELECT p.organization_uuid,
    activities.id
   FROM activities,
    projects p,
    organizations o
  WHERE ((p.organization_uuid = o.uuid) AND (((activities.extra -> 'projectUuid'::text))::text = (('"'::text || p.uuid) || '"'::text)))
UNION
 SELECT (substr((((activities.extra -> 'Organization'::text) -> 'uuid'::text))::text, 1, (length((((activities.extra -> 'Organization'::text) -> 'uuid'::text))::text) - 1)))::uuid AS organization_uuid,
    activities.id
   FROM activities
  WHERE (activities.name = 'organization.created'::text)
  WITH NO DATA;

REFRESH MATERIALIZED VIEW organization_activities;

--
-- Name: orphaned_activities; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE orphaned_activities (
    id bigint,
    job_uuid text
);



--
-- Name: provider_plan_availabilities_and_limits; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE provider_plan_availabilities_and_limits (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    provider_name text NOT NULL,
    provider_plan_id text,
    availability daterange NOT NULL,
    price_per_additional_user text,
    limits json NOT NULL,
    CONSTRAINT price_per_additional_user_fmt CHECK ((price_per_additional_user ~ '^[0-9]{1,}\.[0-9]+\s[A-Z]{3}$|^$'::text))
);



--
-- Name: repositories; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE repositories (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    project_uuid uuid NOT NULL,
    name text NOT NULL,
    url text NOT NULL,
    visible_to membership_type DEFAULT 'member'::membership_type NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    accessible boolean DEFAULT false NOT NULL,
    github_imported boolean DEFAULT false,
    github_login character varying(100),
    github_repo character varying(100),
    archived_at timestamp with time zone,
    metadata json,
    metadata_updated_at timestamp with time zone
);



--
-- Name: repository_credentials; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE repository_credentials (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    repository_uuid uuid NOT NULL,
    type repository_credential_type NOT NULL,
    archived_at timestamp with time zone,
    status repository_credential_status DEFAULT 'pending'::repository_credential_status NOT NULL,
    key bytea DEFAULT gen_random_bytes(56) NOT NULL
);



--
-- Name: schedules; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE schedules (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    job_uuid uuid NOT NULL,
    user_uuid uuid NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    cronspec character varying(100),
    description character varying(200) NOT NULL,
    run_once_at timestamp with time zone,
    location text DEFAULT 'UTC'::text NOT NULL,
    disabled_because text,
    timespec character varying(100),
    disabled schedule_disabled_type,
    parameters json,
    archived_at timestamp with time zone,
    CONSTRAINT cronspec_run_once_at_check CHECK (((cronspec IS NULL) <> (run_once_at IS NULL))),
    CONSTRAINT schedules_location_check CHECK (is_valid_timezone_name((location)::character varying))
);



--
-- Name: secrets; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE secrets (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    environment_uuid uuid NOT NULL,
    type secret_type NOT NULL,
    archived_at timestamp with time zone,
    status secret_status DEFAULT 'pending'::secret_status NOT NULL,
    key bytea DEFAULT gen_random_bytes(56) NOT NULL
);



--
-- Name: sessions; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE sessions (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    user_uuid uuid NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    validated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    expires_at timestamp with time zone DEFAULT date_trunc('day'::text, (timezone('utc'::text, now()) + '15 days'::interval)) NOT NULL,
    logged_out_at timestamp with time zone,
    user_agent text NOT NULL,
    client_address text NOT NULL,
    invalidated_at timestamp without time zone
);



--
-- Name: slack_notifiers; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE slack_notifiers (
    uuid uuid NOT NULL,
    name text NOT NULL,
    url_host text NOT NULL,
    project_uuid uuid NOT NULL,
    webhook_url text NOT NULL,
    archived_at timestamp with time zone
);



--
-- Name: stalled_jobs_2017_02_26_lee; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE stalled_jobs_2017_02_26_lee (
    uuid uuid,
    type operation_type,
    workspace_base_image_uuid uuid,
    repository_uuid uuid,
    time_limit integer,
    exit_status integer,
    created_at timestamp with time zone,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    timed_out_at timestamp with time zone,
    fatal_error text,
    job_uuid uuid,
    failed_at timestamp with time zone,
    archived_at timestamp with time zone,
    parameters json,
    repository_refs json,
    git_logs json,
    status_logs json,
    notifier_uuid uuid,
    notifier_type text,
    canceled_at timestamp without time zone
);



--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE subscriptions (
    uuid uuid NOT NULL,
    user_uuid uuid NOT NULL,
    watchable_uuid uuid NOT NULL,
    watchable_type text NOT NULL,
    event_name text NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    archived_at timestamp with time zone
);



--
-- Name: targets; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE targets (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    type target_type NOT NULL,
    project_uuid uuid NOT NULL,
    environment_uuid uuid NOT NULL,
    url text NOT NULL,
    identifier text,
    secret text,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    accessibe_since timestamp with time zone,
    archived_at timestamp with time zone
);



--
-- Name: user_blocks; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE user_blocks (
    uuid uuid NOT NULL,
    user_uuid uuid,
    reason text NOT NULL,
    valid tstzrange NOT NULL
);



--
-- Name: users; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE users (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    email character varying(254) NOT NULL,
    name text NOT NULL,
    token character varying(64) DEFAULT encode(gen_random_bytes(32), 'hex'::text) NOT NULL,
    totp_secret character varying(64) DEFAULT ''::character varying NOT NULL,
    password_reset_token character varying(64) DEFAULT ''::character varying NOT NULL,
    password_hash character varying(60) DEFAULT ''::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    url_host character varying(100) NOT NULL,
    totp_enabled_at timestamp with time zone,
    gh_username text,
    without_password boolean DEFAULT false,
    signup_parameters json
);



--
-- Name: webhooks; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE webhooks (
    uuid uuid NOT NULL,
    project_uuid uuid NOT NULL,
    archived_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    creator_uuid uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    job_uuid uuid
);



--
-- Name: workspace_base_images; Type: TABLE; Schema: public; Owner: production
--

CREATE TABLE workspace_base_images (
    uuid uuid DEFAULT uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    repository character varying(255) NOT NULL,
    path character varying(100) DEFAULT './'::character varying NOT NULL,
    ref character varying(100) DEFAULT 'master'::character varying NOT NULL,
    type workspace_base_image_type DEFAULT 'container'::workspace_base_image_type NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);



--
-- Name: activities id; Type: DEFAULT; Schema: public; Owner: production
--

ALTER TABLE ONLY activities ALTER COLUMN id SET DEFAULT nextval('activities_id_seq'::regclass);


--
-- Name: activities activities_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY activities
    ADD CONSTRAINT activities_pkey PRIMARY KEY (id);


--
-- Name: billing_events billing_events_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY billing_events
    ADD CONSTRAINT billing_events_pkey PRIMARY KEY (uuid);


--
-- Name: deliveries deliveries_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY deliveries
    ADD CONSTRAINT deliveries_pkey PRIMARY KEY (uuid);


--
-- Name: email_notifiers email_notifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY email_notifiers
    ADD CONSTRAINT email_notifiers_pkey PRIMARY KEY (uuid);


--
-- Name: environments environments_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY environments
    ADD CONSTRAINT environments_pkey PRIMARY KEY (uuid);


--
-- Name: git_triggers git_triggers_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY git_triggers
    ADD CONSTRAINT git_triggers_pkey PRIMARY KEY (uuid);


--
-- Name: invitations invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY invitations
    ADD CONSTRAINT invitations_pkey PRIMARY KEY (uuid);


--
-- Name: job_notifiers job_notifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY job_notifiers
    ADD CONSTRAINT job_notifiers_pkey PRIMARY KEY (uuid);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (uuid);


--
-- Name: notification_rules notification_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY notification_rules
    ADD CONSTRAINT notification_rules_pkey PRIMARY KEY (uuid);


--
-- Name: oauth_tokens oauth_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY oauth_tokens
    ADD CONSTRAINT oauth_tokens_pkey PRIMARY KEY (uuid);


--
-- Name: operations operations_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY operations
    ADD CONSTRAINT operations_pkey PRIMARY KEY (uuid);


--
-- Name: organizations organisations_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY organizations
    ADD CONSTRAINT organisations_pkey PRIMARY KEY (uuid);


--
-- Name: project_memberships project_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY project_memberships
    ADD CONSTRAINT project_memberships_pkey PRIMARY KEY (uuid);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (uuid);


--
-- Name: provider_plan_availabilities_and_limits provider_plan_availabilities__provider_plan_id_availabilit_excl; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY provider_plan_availabilities_and_limits
    ADD CONSTRAINT provider_plan_availabilities__provider_plan_id_availabilit_excl EXCLUDE USING gist (provider_plan_id WITH =, availability WITH &&);


--
-- Name: provider_plan_availabilities_and_limits provider_plan_availabilities_and_limits_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY provider_plan_availabilities_and_limits
    ADD CONSTRAINT provider_plan_availabilities_and_limits_pkey PRIMARY KEY (uuid);


--
-- Name: repositories repositories_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY repositories
    ADD CONSTRAINT repositories_pkey PRIMARY KEY (uuid);


--
-- Name: repository_credentials repository_credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY repository_credentials
    ADD CONSTRAINT repository_credentials_pkey PRIMARY KEY (uuid);


--
-- Name: schedules schedules_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY schedules
    ADD CONSTRAINT schedules_pkey PRIMARY KEY (uuid);


--
-- Name: secrets secrets_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (uuid);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (uuid);


--
-- Name: slack_notifiers slack_notifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY slack_notifiers
    ADD CONSTRAINT slack_notifiers_pkey PRIMARY KEY (uuid);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (uuid);


--
-- Name: targets targets_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY targets
    ADD CONSTRAINT targets_pkey PRIMARY KEY (uuid);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (uuid);


--
-- Name: user_blocks user_blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY user_blocks
    ADD CONSTRAINT user_blocks_pkey PRIMARY KEY (uuid);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (uuid);


--
-- Name: users users_token_key; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_token_key UNIQUE (token);


--
-- Name: webhooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (uuid);


--
-- Name: workspace_base_images workspace_base_images_pkey; Type: CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY workspace_base_images
    ADD CONSTRAINT workspace_base_images_pkey PRIMARY KEY (uuid);


--
-- Name: activities_archive_name; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX activities_archive_name ON activities USING btree (name);


--
-- Name: activities_context_user_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX activities_context_user_uuid ON activities USING btree (context_user_uuid);


--
-- Name: activities_name; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX activities_name ON activities USING btree (name);


--
-- Name: activities_occurred_on; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX activities_occurred_on ON activities USING btree (occurred_on DESC);


--
-- Name: activity_archove_id; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX activity_archove_id ON activities_archive USING btree (id);


--
-- Name: billing_events_organization_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX billing_events_organization_uuid ON billing_events USING btree (organization_uuid);


--
-- Name: enabled_schedules_index; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX enabled_schedules_index ON schedules USING btree (disabled) WHERE (disabled IS NULL);


--
-- Name: environments_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX environments_archived_at ON environments USING btree (archived_at);


--
-- Name: environments_by_project_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX environments_by_project_uuid ON environments USING btree (archived_at, project_uuid);


--
-- Name: jobs_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX jobs_archived_at ON jobs USING btree (archived_at);


--
-- Name: jobs_guard_no_duplicate_jobs; Type: INDEX; Schema: public; Owner: production
--

CREATE UNIQUE INDEX jobs_guard_no_duplicate_jobs ON jobs USING btree (task_uuid, environment_uuid) WHERE (timezone('UTC'::text, created_at) > date(timezone('UTC'::text, '2016-09-06 00:00:00+00'::timestamp with time zone)));


--
-- Name: jobs_is_active_user_job; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX jobs_is_active_user_job ON jobs USING btree (archived_at, name) WHERE is_active_user_job(archived_at, ((name)::text)::character varying);


--
-- Name: no_duplicate_organisation_memberships; Type: INDEX; Schema: public; Owner: production
--

CREATE UNIQUE INDEX no_duplicate_organisation_memberships ON organization_memberships USING btree (organization_uuid, user_uuid);


--
-- Name: notifiction_rules_job_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX notifiction_rules_job_uuid ON notification_rules USING btree (archived_at, job_uuid);


--
-- Name: one_token_per_provider; Type: INDEX; Schema: public; Owner: production
--

CREATE UNIQUE INDEX one_token_per_provider ON oauth_tokens USING btree (provider, user_uuid);


--
-- Name: operations_at_fields; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX operations_at_fields ON operations USING btree (started_at, canceled_at, timed_out_at, failed_at, finished_at, archived_at);


--
-- Name: operations_created_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX operations_created_at ON operations USING btree (created_at);


--
-- Name: operations_end_fields_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX operations_end_fields_at ON operations USING btree (canceled_at, timed_out_at, failed_at, finished_at, archived_at);


--
-- Name: operations_job_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX operations_job_uuid ON operations USING btree (job_uuid);


--
-- Name: operations_started_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX operations_started_at ON operations USING btree (created_at);


--
-- Name: operations_type_job; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX operations_type_job ON operations USING btree (type) WHERE is_job_operation(type);


--
-- Name: organization_memberships_type; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX organization_memberships_type ON organization_memberships USING btree (type);


--
-- Name: organizations_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX organizations_archived_at ON organizations USING btree (archived_at);


--
-- Name: projects_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX projects_archived_at ON projects USING btree (archived_at);


--
-- Name: projects_public; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX projects_public ON projects USING btree (public);


--
-- Name: provider_plan_availabilities_and_limits_availability; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX provider_plan_availabilities_and_limits_availability ON provider_plan_availabilities_and_limits USING gist (availability);


--
-- Name: repositories_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX repositories_archived_at ON repositories USING btree (archived_at);


--
-- Name: repositories_by_project_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX repositories_by_project_uuid ON repositories USING btree (archived_at, project_uuid);


--
-- Name: repository_credentials_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX repository_credentials_archived_at ON repository_credentials USING btree (archived_at);


--
-- Name: repository_credentials_repository_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX repository_credentials_repository_uuid ON repository_credentials USING btree (repository_uuid);


--
-- Name: secrets_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX secrets_archived_at ON secrets USING btree (archived_at);


--
-- Name: secrets_environment_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX secrets_environment_uuid ON secrets USING btree (environment_uuid);


--
-- Name: subscriptions_by_watchable_event_user; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX subscriptions_by_watchable_event_user ON subscriptions USING btree (watchable_uuid, event_name, user_uuid);


--
-- Name: tasks_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX tasks_archived_at ON tasks USING btree (archived_at);


--
-- Name: user_blocks_user_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX user_blocks_user_uuid ON user_blocks USING btree (user_uuid);


--
-- Name: users_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX users_archived_at ON users USING btree (totp_secret);


--
-- Name: webhooks_archived_at; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX webhooks_archived_at ON webhooks USING btree (archived_at);


--
-- Name: webhooks_project_uuid; Type: INDEX; Schema: public; Owner: production
--

CREATE INDEX webhooks_project_uuid ON webhooks USING btree (project_uuid);


--
-- Name: webhooks_unique_slugs; Type: INDEX; Schema: public; Owner: production
--

CREATE UNIQUE INDEX webhooks_unique_slugs ON webhooks USING btree (slug);


--
-- Name: users broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: projects broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON projects FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: organizations broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON organizations FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: organization_memberships broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON organization_memberships FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: repositories broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON repositories FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: operations broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON operations FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: targets broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON targets FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: environments broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON environments FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: tasks broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON tasks FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: jobs broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON jobs FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: schedules broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON schedules FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: project_memberships broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON project_memberships FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: invitations broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON invitations FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: subscriptions broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON subscriptions FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: repository_credentials broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON repository_credentials FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: secrets broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON secrets FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: webhooks broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON webhooks FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: user_blocks broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON user_blocks FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: git_triggers broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON git_triggers FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: email_notifiers broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON email_notifiers FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: notification_rules broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON notification_rules FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: job_notifiers broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON job_notifiers FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: slack_notifiers broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON slack_notifiers FOR EACH ROW EXECUTE PROCEDURE broadcast_change();


--
-- Name: sessions broadcast_change; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_change AFTER UPDATE ON sessions FOR EACH ROW EXECUTE PROCEDURE broadcast_session_change();


--
-- Name: users broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON users FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: sessions broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON sessions FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: projects broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON projects FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: organizations broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON organizations FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: organization_memberships broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON organization_memberships FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: repositories broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON repositories FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: operations broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON operations FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: targets broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON targets FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: environments broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON environments FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: tasks broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON tasks FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: jobs broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON jobs FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: schedules broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON schedules FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: project_memberships broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER UPDATE ON project_memberships FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: invitations broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON invitations FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: subscriptions broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON subscriptions FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: repository_credentials broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON repository_credentials FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: secrets broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON secrets FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: webhooks broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON webhooks FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: deliveries broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON deliveries FOR EACH ROW EXECUTE PROCEDURE broadcast_create_uuid();


--
-- Name: user_blocks broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON user_blocks FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: billing_events broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON billing_events FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: activities broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON activities FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: git_triggers broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON git_triggers FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: email_notifiers broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON email_notifiers FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: notification_rules broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON notification_rules FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: job_notifiers broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON job_notifiers FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: slack_notifiers broadcast_create; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER broadcast_create AFTER INSERT ON slack_notifiers FOR EACH ROW EXECUTE PROCEDURE broadcast_create();


--
-- Name: environments cascade_environment_archival; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER cascade_environment_archival AFTER UPDATE OF archived_at ON environments FOR EACH ROW EXECUTE PROCEDURE cascade_environment_archival();


--
-- Name: jobs cascade_job_archival; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER cascade_job_archival AFTER UPDATE OF archived_at ON jobs FOR EACH ROW EXECUTE PROCEDURE cascade_job_archival();


--
-- Name: organizations cascade_organization_archival; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER cascade_organization_archival AFTER UPDATE OF archived_at ON organizations FOR EACH ROW EXECUTE PROCEDURE cascade_organization_archival();


--
-- Name: projects cascade_project_archival; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER cascade_project_archival AFTER UPDATE OF archived_at ON projects FOR EACH ROW EXECUTE PROCEDURE cascade_project_archival();


--
-- Name: repositories cascade_repository_archival; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER cascade_repository_archival AFTER UPDATE OF archived_at ON repositories FOR EACH ROW EXECUTE PROCEDURE cascade_repository_archival();


--
-- Name: tasks cascade_task_archival; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER cascade_task_archival AFTER UPDATE OF archived_at ON tasks FOR EACH ROW EXECUTE PROCEDURE cascade_task_archival();


--
-- Name: webhooks cascade_webhook_archival; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER cascade_webhook_archival AFTER UPDATE OF archived_at ON webhooks FOR EACH ROW EXECUTE PROCEDURE cascade_webhook_archival();


--
-- Name: operations notify_new_operation; Type: TRIGGER; Schema: public; Owner: production
--

CREATE TRIGGER notify_new_operation AFTER INSERT OR UPDATE ON operations FOR EACH STATEMENT EXECUTE PROCEDURE notify_new_operation();


--
-- Name: deliveries deliveries_webhook_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY deliveries
    ADD CONSTRAINT deliveries_webhook_uuid_fkey FOREIGN KEY (webhook_uuid) REFERENCES webhooks(uuid);


--
-- Name: email_notifiers email_notifiers_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY email_notifiers
    ADD CONSTRAINT email_notifiers_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: environments environments_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY environments
    ADD CONSTRAINT environments_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: git_triggers git_triggers_creator_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY git_triggers
    ADD CONSTRAINT git_triggers_creator_uuid_fkey FOREIGN KEY (creator_uuid) REFERENCES users(uuid);


--
-- Name: git_triggers git_triggers_job_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY git_triggers
    ADD CONSTRAINT git_triggers_job_uuid_fkey FOREIGN KEY (job_uuid) REFERENCES jobs(uuid);


--
-- Name: git_triggers git_triggers_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY git_triggers
    ADD CONSTRAINT git_triggers_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: git_triggers git_triggers_repository_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY git_triggers
    ADD CONSTRAINT git_triggers_repository_uuid_fkey FOREIGN KEY (repository_uuid) REFERENCES repositories(uuid);


--
-- Name: invitations invitations_creator_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY invitations
    ADD CONSTRAINT invitations_creator_uuid_fkey FOREIGN KEY (creator_uuid) REFERENCES users(uuid);


--
-- Name: invitations invitations_organisation_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY invitations
    ADD CONSTRAINT invitations_organisation_uuid_fkey FOREIGN KEY (organization_uuid) REFERENCES organizations(uuid);


--
-- Name: invitations invitations_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY invitations
    ADD CONSTRAINT invitations_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: job_notifiers job_notifiers_job_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY job_notifiers
    ADD CONSTRAINT job_notifiers_job_uuid_fkey FOREIGN KEY (job_uuid) REFERENCES jobs(uuid);


--
-- Name: job_notifiers job_notifiers_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY job_notifiers
    ADD CONSTRAINT job_notifiers_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: jobs jobs_environment_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT jobs_environment_uuid_fkey FOREIGN KEY (environment_uuid) REFERENCES environments(uuid);


--
-- Name: jobs jobs_task_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY jobs
    ADD CONSTRAINT jobs_task_uuid_fkey FOREIGN KEY (task_uuid) REFERENCES tasks(uuid);


--
-- Name: notification_rules notification_rules_creator_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY notification_rules
    ADD CONSTRAINT notification_rules_creator_uuid_fkey FOREIGN KEY (creator_uuid) REFERENCES users(uuid);


--
-- Name: notification_rules notification_rules_job_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY notification_rules
    ADD CONSTRAINT notification_rules_job_uuid_fkey FOREIGN KEY (job_uuid) REFERENCES jobs(uuid);


--
-- Name: notification_rules notification_rules_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY notification_rules
    ADD CONSTRAINT notification_rules_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: oauth_tokens oauth_tokens_user_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY oauth_tokens
    ADD CONSTRAINT oauth_tokens_user_uuid_fkey FOREIGN KEY (user_uuid) REFERENCES users(uuid);


--
-- Name: operations operations_job_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY operations
    ADD CONSTRAINT operations_job_uuid_fkey FOREIGN KEY (job_uuid) REFERENCES jobs(uuid);


--
-- Name: operations operations_repository_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY operations
    ADD CONSTRAINT operations_repository_uuid_fkey FOREIGN KEY (repository_uuid) REFERENCES repositories(uuid);


--
-- Name: operations operations_workspace_base_image_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY operations
    ADD CONSTRAINT operations_workspace_base_image_uuid_fkey FOREIGN KEY (workspace_base_image_uuid) REFERENCES workspace_base_images(uuid);


--
-- Name: organization_memberships organisation_memberships_organisation_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY organization_memberships
    ADD CONSTRAINT organisation_memberships_organisation_uuid_fkey FOREIGN KEY (organization_uuid) REFERENCES organizations(uuid);


--
-- Name: organization_memberships organisation_memberships_user_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY organization_memberships
    ADD CONSTRAINT organisation_memberships_user_uuid_fkey FOREIGN KEY (user_uuid) REFERENCES users(uuid);


--
-- Name: project_memberships project_memberships_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY project_memberships
    ADD CONSTRAINT project_memberships_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: project_memberships project_memberships_user_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY project_memberships
    ADD CONSTRAINT project_memberships_user_uuid_fkey FOREIGN KEY (user_uuid) REFERENCES users(uuid);


--
-- Name: projects projects_organisation_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY projects
    ADD CONSTRAINT projects_organisation_uuid_fkey FOREIGN KEY (organization_uuid) REFERENCES organizations(uuid);


--
-- Name: repositories repositories_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY repositories
    ADD CONSTRAINT repositories_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: repository_credentials repository_credentials_repository_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY repository_credentials
    ADD CONSTRAINT repository_credentials_repository_uuid_fkey FOREIGN KEY (repository_uuid) REFERENCES repositories(uuid);


--
-- Name: schedules schedules_job_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY schedules
    ADD CONSTRAINT schedules_job_uuid_fkey FOREIGN KEY (job_uuid) REFERENCES jobs(uuid);


--
-- Name: schedules schedules_user_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY schedules
    ADD CONSTRAINT schedules_user_uuid_fkey FOREIGN KEY (user_uuid) REFERENCES users(uuid);


--
-- Name: secrets secrets_environment_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY secrets
    ADD CONSTRAINT secrets_environment_uuid_fkey FOREIGN KEY (environment_uuid) REFERENCES environments(uuid);


--
-- Name: sessions sessions_user_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY sessions
    ADD CONSTRAINT sessions_user_uuid_fkey FOREIGN KEY (user_uuid) REFERENCES users(uuid);


--
-- Name: slack_notifiers slack_notifiers_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY slack_notifiers
    ADD CONSTRAINT slack_notifiers_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: subscriptions subscriptions_user_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY subscriptions
    ADD CONSTRAINT subscriptions_user_uuid_fkey FOREIGN KEY (user_uuid) REFERENCES users(uuid);


--
-- Name: targets targets_environment_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY targets
    ADD CONSTRAINT targets_environment_uuid_fkey FOREIGN KEY (environment_uuid) REFERENCES environments(uuid);


--
-- Name: targets targets_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY targets
    ADD CONSTRAINT targets_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: tasks tasks_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY tasks
    ADD CONSTRAINT tasks_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- Name: user_blocks user_blocks_user_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY user_blocks
    ADD CONSTRAINT user_blocks_user_uuid_fkey FOREIGN KEY (user_uuid) REFERENCES users(uuid);


--
-- Name: webhooks webhooks_creator_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY webhooks
    ADD CONSTRAINT webhooks_creator_uuid_fkey FOREIGN KEY (creator_uuid) REFERENCES users(uuid);


--
-- Name: webhooks webhooks_job_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY webhooks
    ADD CONSTRAINT webhooks_job_uuid_fkey FOREIGN KEY (job_uuid) REFERENCES jobs(uuid);


--
-- Name: webhooks webhooks_project_uuid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: production
--

ALTER TABLE ONLY webhooks
    ADD CONSTRAINT webhooks_project_uuid_fkey FOREIGN KEY (project_uuid) REFERENCES projects(uuid);


--
-- PostgreSQL database dump complete
--

-- +migrate StatementEnd

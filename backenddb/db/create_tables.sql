/*
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/

--
-- PostgreSQL database dump
--

-- Dumped from database version 13.1
-- Dumped by pg_dump version 13.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner:
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


SET search_path = public, pg_catalog;

--
-- Name: on_update_current_timestamp_last_updated(); Type: FUNCTION; Schema: public; Owner: traffic_vault
--

CREATE OR REPLACE FUNCTION on_update_current_timestamp_last_updated() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.last_updated = now();
RETURN NEW;
END;
$$;


ALTER FUNCTION on_update_current_timestamp_last_updated() OWNER TO traffic_ops;

SET default_tablespace = '';

--
-- Name: foo; Type: TABLE; Schema: public; Owner: traffic_ops
--

CREATE TABLE IF NOT EXISTS foo (
                                    id bigserial,
                                    name text NOT NULL,
                                    last_updated timestamp with time zone DEFAULT now() NOT NULL,
                                    CONSTRAINT idx_89491_primary PRIMARY KEY (id)
    );

ALTER TABLE foo OWNER TO traffic_ops;

--
-- Name: cdn_id_seq; Type: SEQUENCE; Schema: public; Owner: traffic_ops
--

CREATE SEQUENCE IF NOT EXISTS foo_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE foo_id_seq OWNER TO traffic_ops;

ALTER SEQUENCE foo_id_seq OWNED BY foo.id;
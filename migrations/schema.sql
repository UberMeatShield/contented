--
-- PostgreSQL database dump
--

-- Dumped from database version 10.7
-- Dumped by pg_dump version 12.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

--
-- Name: containers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.containers (
    id uuid NOT NULL,
    total integer DEFAULT 0 NOT NULL,
    path character varying(255) DEFAULT ''::character varying NOT NULL,
    name character varying(255) DEFAULT ''::character varying NOT NULL,
    active boolean DEFAULT TRUE,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.containers OWNER TO postgres;

--
-- Name: media_containers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.media_containers (
    id uuid NOT NULL,
    src character varying(255) DEFAULT ''::character varying NOT NULL,
    type character varying(255) DEFAULT ''::character varying NOT NULL,
    preview character varying(255) DEFAULT ''::character varying NOT NULL,
    container_id uuid,
    active boolean DEFAULT TRUE,
    idx integer DEFAULT 0 NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.media_containers OWNER TO postgres;

--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO postgres;

--
-- Name: containers containers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.containers
    ADD CONSTRAINT containers_pkey PRIMARY KEY (id);


--
-- Name: media_containers media_containers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.media_containers
    ADD CONSTRAINT media_containers_pkey PRIMARY KEY (id);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: media_containers media_containers_container_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.media_containers
    ADD CONSTRAINT media_containers_container_id_fkey FOREIGN KEY (container_id) REFERENCES public.containers(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


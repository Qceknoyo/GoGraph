--
-- PostgreSQL database dump
--

-- \restrict fVNuQEeAPICl4W1PnslEMcOU0lIjDRGWgdNQEwDrqbbfgS4YybVWCTS47J1ziQv

-- Dumped from database version 18.4
-- Dumped by pg_dump version 18.4

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

SET default_table_access_method = heap;

--
-- Name: experiments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.experiments (
    id integer NOT NULL,
    title character varying(255) NOT NULL,
    material character varying(100) NOT NULL,
    operator character varying(50) DEFAULT 'Admin'::character varying,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    user_id uuid
);


--
-- Name: experiments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.experiments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: experiments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.experiments_id_seq OWNED BY public.experiments.id;


--
-- Name: measurements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.measurements (
    id bigint NOT NULL,
    experiment_id integer NOT NULL,
    strain double precision NOT NULL,
    stress double precision NOT NULL,
    extra_data jsonb DEFAULT '{}'::jsonb,
    "timestamp" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: measurements_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.measurements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: measurements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.measurements_id_seq OWNED BY public.measurements.id;


--
-- Name: sync_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sync_tokens (
    token text NOT NULL,
    user_id uuid,
    expires_at timestamp without time zone DEFAULT (now() + '00:15:00'::interval)
);


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_sessions (
    token character varying(64) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    user_id uuid,
    expires_at timestamp without time zone DEFAULT (now() + '30 days'::interval)
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    last_seen timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: experiments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.experiments ALTER COLUMN id SET DEFAULT nextval('public.experiments_id_seq'::regclass);


--
-- Name: measurements id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.measurements ALTER COLUMN id SET DEFAULT nextval('public.measurements_id_seq'::regclass);


--
-- Name: experiments experiments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.experiments
    ADD CONSTRAINT experiments_pkey PRIMARY KEY (id);


--
-- Name: measurements measurements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.measurements
    ADD CONSTRAINT measurements_pkey PRIMARY KEY (id);


--
-- Name: sync_tokens sync_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sync_tokens
    ADD CONSTRAINT sync_tokens_pkey PRIMARY KEY (token);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (token);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_measurements_experiment_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_measurements_experiment_id ON public.measurements USING btree (experiment_id);


--
-- Name: experiments experiments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.experiments
    ADD CONSTRAINT experiments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: measurements measurements_experiment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.measurements
    ADD CONSTRAINT measurements_experiment_id_fkey FOREIGN KEY (experiment_id) REFERENCES public.experiments(id) ON DELETE CASCADE;


--
-- Name: sync_tokens sync_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sync_tokens
    ADD CONSTRAINT sync_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

-- \unrestrict fVNuQEeAPICl4W1PnslEMcOU0lIjDRGWgdNQEwDrqbbfgS4YybVWCTS47J1ziQv


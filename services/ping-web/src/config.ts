// Baked in at build time (see Dockerfile ARG / deploy.yml build-args)
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

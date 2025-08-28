declare module "cloudflare:test" {
  interface ProvidedEnv {
    KV_NAMESPACE: KVNamespace
  }
  // ...or if you have an existing `Env` type...
  interface ProvidedEnv extends Env {}
}
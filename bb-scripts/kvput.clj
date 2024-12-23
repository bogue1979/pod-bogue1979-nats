#!/usr/bin/env bb
(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost"
   :nkey (slurp (str (System/getenv "HOME") "/.config/nats/local.seed")),
   :msg "_"
   :bucket "first-bucket"})

(prn (nats/kvput (merge opts {:key "foo" :value "bar"})))
(prn (nats/kvget (merge opts {:key "foo"})))

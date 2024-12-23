#!/usr/bin/env bb
(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost"
   :nkey (slurp (str (System/getenv "HOME") "/.config/nats/local.seed")),
   :msg "_"
   :bucket "first-bucket"})

(nats/kvwatchbucket (fn [msg] (prn msg)) opts)
(deref (promise))

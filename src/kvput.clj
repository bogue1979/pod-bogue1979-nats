#!/usr/bin/env bb
(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost" 
   :nkey "SUABUVFNUZOLGESMUFWGGA766NRZUDCJFDM2JIQSYNVP5AIPWTC4GBMAZY",
   :msg "_"
   :bucket "first-bucket"})

(nats/kvput (merge opts {:key "foo" :value "foo bas"}))

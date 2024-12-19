#!/usr/bin/env bb
(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost"
   :nkey (slurp "/home/danny/.config/nats/local.seed"),
   :msg "_"
   :bucket "first-bucket"})

;;(nats/kvput {:host "localhost"
;;             :nkey (slurp "/home/danny/.config/nats/local.seed"),
;;             :msg "_"
;;             :key "bat"})

(prn (nats/kvput (merge opts {:key "foo" :value "what_value"})))
(prn (nats/kvget (merge opts {:key "foo"})))
(prn (nats/kvget (merge opts {:key "non"})))
(prn (nats/kvget (merge opts {:key "foo"})))

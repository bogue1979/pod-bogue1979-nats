(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
;(some? (find-ns 'pod.bogue1979.nats))
(require '[pod.bogue1979.nats :as nats])

(defn pub [h n s m] (nats/publish {:host h, :nkey n, :subject s, :msg m}))


(def opts
  {:host "localhost"
   :nkey (slurp "/home/danny/.config/nats/local.seed"),
   :msg "foo"
   :subject "metrics.bay"})

(prn (nats/publish opts))


(prn (nats/publish {:host "localhost"
                    :nkey (slurp "/home/danny/.config/nats/local.seed")
                    :msg "bar"
                    :subject "not.allowed"}))

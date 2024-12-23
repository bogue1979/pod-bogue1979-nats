(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
;(some? (find-ns 'pod.bogue1979.nats))
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost"
   :nkey (slurp (str (System/getenv "HOME") "/.config/nats/local.seed")),
   :msg "{\"foo\": \"baz\"}"
   :subject "metrics.test"})

(prn (nats/publish opts))
;(prn (nats/publish (merge opts {:subject "not.allowed"})))

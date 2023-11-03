(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")

(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost" 
   :nkey "SUABUVFNUZOLGESMUFWGGA766NRZUDCJFDM2JIQSYNVP5AIPWTC4GBMAZY"
   :subject "metrics.qa-server" 
   :timeout_seconds 10})

(println (try
       (nats/request (merge opts {:msg "Das ist ein Test"}))
       (catch Exception e (str "Error: " (.getMessage e)))))

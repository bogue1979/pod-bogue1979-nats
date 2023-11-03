#!/usr/bin/env bb

(require '[clojure.core.async :as a]
         '[babashka.pods :as pods]
         ;'[clojure.java.shell :refer [ sh]]
         ;'[cheshire.core :as json :refer [parse-string]]
         )
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost" 
   :nkey "SUABUVFNUZOLGESMUFWGGA766NRZUDCJFDM2JIQSYNVP5AIPWTC4GBMAZY",
   :bucket "first-bucket"})

(defn runsubscriber
  [ch]
  (.addShutdownHook (Runtime/getRuntime)
                    (Thread. #(do
                                (println  "SHUTDOWN")
                                (a/close! ch))))
  (println "Start subscriber")
  (a/go (nats/kvwatchbucket (fn [msg] (a/>! ch msg)) opts)))


(let [ch (a/chan)]
  (runsubscriber ch)
  (while true
    (prn (a/<!! ch))))



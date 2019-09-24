CREATE TABLE job (
  id INT NOT NULL AUTO_INCREMENT, 
  event_type VARCHAR(20) NOT NULL,
  resource TEXT,
  create_time DATETIME,
  PRIMARY KEY(id)
);

CREATE TABLE job_pod (
  job_id INT NOT NULL,
  pod   VARCHAR(255) NOT NULL,
  FOREIGN KEY (job_id) references job(id)
);

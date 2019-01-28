### yams-dav-syncher

Script to upload images from yapo's DAV server to yams.

##### Requirements
- PostgresDB [9.0 >]
- The maximum number of open files / file descriptors:

```
 > 5000  - availables for sync process
 > 10000 - availables for deleteall process
 ```

 In accord to your requirement use:
```
 ulimit -n 
 
 ulimit -n 5000
```

##### how to use

 - Clone the repository
 - Get your rsa key and include that in `./private-key.rsa`
 - Use `./scripts/commands/vars.mk` to configurate the following params:

```
export DATABASE_NAME=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=pgdb
export DATABASE_PASSWORD=postgres

export YAMS_MGMT_URL=https://mgmt-us-east-1-yams.schibsted.com/api/v1
export YAMS_TENTAND_ID=[your-tentant]
export YAMS_DOMAIN_ID=[your-domain]
export YAMS_BUCKET_ID=[your-bucket]
export YAMS_ACCESS_KEY_ID=[your-access-key]
export YAMS_PRIVATE_KEY=${PWD}/file-key.rsa
export YAMS_UPLOAD_LIMIT=0 # 0 means no limits 
export YAMS_MAX_CONCURRENT_CONN=100
export YAMS_TiMEOUT=30

export BANDWIDTH_PROXY_LIMIT=500 # kbps

export IMAGES_PATH=/images/uploads/
```

 - `make compress` to generate the ready-to-deploy binaries compressed in `./output/`
 -  Upload the tar.gz file and decompress it in your dav server
 -  In dav server you can edit `script/commands/vars.mk` modify config vars
 -  type `make sync` or `make sync&` (detached mode) to do:
    1) Generate a file sorted-list with images of `IMAGES_PATH` sorted by date 
    2) Upload each image of the list using concurrency
    3) In case of error then mark in DB the retry to upload in the next script execution
    4) Mark in DB the date of the last synchronized image, thus with a new `make sync` the process will start from this date (skipping older images from the sorted-list).
    Note: with each execution a new images sorted-list will be generated but also will be deleted when the execution is done.

###### Other commnads

- `make list` to list the images in yams bucket
- `make deleteall` to delete everything stored in yams bucket
- `make markslist` to get a list with all synchronization mark ordered by newer to older
- `make reset` deletes the last synchronization mark

- `make sync&` to execute sync process in detached mode
- `make deleteall&` to delete everything stored in yams bucket in detached mode
- `make list&` to list the images in yams bucket in detached mode

NOTE: Make deleteall & make list use yams pagination to work

###### Monitoring

By default, when the process starts prometheus metrics are exposed in `http://HOST:8877/metrics`

###### Main Process

![image](https://confluence.schibsted.io/rest/gliffy/1.0/embeddedDiagrams/710380ec-5d52-4455-8c9b-77d70e60c4a7.png)


###### Concurrent Send Sub-Process
![image](https://confluence.schibsted.io/rest/gliffy/1.0/embeddedDiagrams/b732a9dd-00f6-46d9-8054-ebb2a653c6e7.png)


###### Error Control Sub-process
![image](https://confluence.schibsted.io/rest/gliffy/1.0/embeddedDiagrams/482d854e-cf28-401e-9730-d2f7bf429f25.png)

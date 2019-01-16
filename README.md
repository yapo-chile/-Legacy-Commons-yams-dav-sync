### yams-dav-syncher

Script to upload images from yapo's DAV server to yams.

##### Requirements
- PostgresDB [9.0 >]

##### how to use

 - Clone the repository
 - Get your .rsa key and set the location ajusting the config in `./scripts/commands/vars.mk`
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
 - `cd ./output/yams-dav-syncher/` and type `make sync` to do:
    1) Generate a file sorted-list with images of `IMAGES_PATH` sorted by date 
    2) Upload each image of the list using concurrency
    3) In case of error then mark in DB the retry to upload in the next script execution
    4) Mark in DB the date of the last synchronized image, thus with a new `make sync` the process will start from this date (skipping older images from the sorted-list).
    Note: with each execution a new images sorted-list will be generated but also will be deleted when the execution is done.

###### Other commnads

- `make list` to list the images in yams bucket
- `make deleteall` to delete everything stored in yams bucket

NOTE: Do not use this commnads when you have too many images in yams bucket. They are only for test purpose.


###### Main Process

![image](https://confluence.schibsted.io/rest/gliffy/1.0/embeddedDiagrams/710380ec-5d52-4455-8c9b-77d70e60c4a7.png)


###### Concurrent Send Sub-Process
![image](https://confluence.schibsted.io/rest/gliffy/1.0/embeddedDiagrams/b732a9dd-00f6-46d9-8054-ebb2a653c6e7.png)


###### Error Control Sub-process
![image](https://confluence.schibsted.io/rest/gliffy/1.0/embeddedDiagrams/482d854e-cf28-401e-9730-d2f7bf429f25.png)

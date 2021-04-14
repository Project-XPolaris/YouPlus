# YouPlus
![](https://img.shields.io/badge/Project-Project%20Polaris-green) 
![](https://img.shields.io/badge/Project-YouPlus-green) 
![](https://img.shields.io/badge/Version-1.0.0-yellow) 
![](https://img.shields.io/badge/Plantform-linux-red)

zfs + samba  = YouPlus

## Feature

- System info

disk info\os info...etc
- ZFS

manage zfs pool(create,remove,use zfs pool as storage...)

- Share Folder(Samba)

manage samba folder (create,remove,edit account access...)

- User

manage user for share folder

## requirements
1. samba service
2. ZFS manage tool (zpool)
3. [YouSMB service](https://github.com/Project-XPolaris/YouSMB)

## config
```json
{
  "addr": ":8999",
  "api_key": "youplus_api_key",
  "yousmb_addr": "http://localhost:8200",
  "fstab": "/etc/fstab"
}
```
`addr` - listen port

`api_key` - key for generate token

`yousmb_addr` - yousmb service address

`fstab` - fstab location

## GUI
[YouPlus Web dashboard](https://github.com/Project-XPolaris/YouPlus-Web) for YouPlus service

### 🔗Links
- [🔨YouPlus Web dashboard](https://github.com/Project-XPolaris/YouPlus-Web)
- [🔨YouSMB ](https://github.com/Project-XPolaris/YouSMB)
- [⭐️Project Polaris](https://github.com/Project-XPolaris)
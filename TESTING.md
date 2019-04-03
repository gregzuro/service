## Heartbeat Test ##

The heartbeat test involves the simplest complete cluster configuration:

 - a master node
 - a slave
 - a plugin
 - a device
 
### Master, Slave, and Plugin Setup

**Ensure you have followed the Readme's setup guide**

To setup your environment cd to the switch directory and run the following commands:
```
❯ make local-cluster-up
❯ make local-slave-start
❯ make local-plugin-up
```

Now we must load data into the plugin from a file (TODO rename "email_file_path").  
```
❯ cd ..
❯ GOPATH=$PWD/.gopath go run ./.gopath/src/github.com/gregzuro/service/plugin/cmd/rpctestingsupport/testclient.go \
-server_addr=127.0.0.1:10011 \
-email_file_path=/root/go/src/github.com/gregzuro/service/email.json
```

The "server_addr" argument can be found by running the following command (look for IPAddress, use 10011 for port):
```
❯ docker inspect sgms_plugin001
```

The "email_file_path" is a configuration for plugin events and landmarks. Here is an example:

```json
{"branches":[{"branch":3,"decisions":[{"Inbranch":3,"Sequence":1,"trueBranch":1,"expression":"state","Operation":"==","Value":"OutOfArea"},{"Inbranch":3,"Sequence":2,"trueBranch":2,"expression":"state","Operation":"==","Value":"InArea"},{"Inbranch":3,"Sequence":3,"trueBranch":9,"expression":"state","Operation":"==","Value":"EnterArea"},{"Inbranch":3,"Sequence":4,"trueBranch":8,"expression":"state","Operation":"==","Value":"ExitArea"}],"branches":[{"branch":1,"decisions":[{"Inbranch":1,"Sequence":5,"trueBranch":4,"expression":"SlideingWindowGPSInArea(gc,5)==3 ||  SlideingWindowWifiInArea(gc,5)==3 || SlideingWindowIBeaconInArea(gc,5)==3","Operation":"==","Value":"true"},{"Inbranch":1,"Sequence":6,"trueBranch":6,"expression":"SlideingWindowGPSInArea(gc,5)==3 ||  SlideingWindowWifiInArea(gc,5)==3 || SlideingWindowIBeaconInArea(gc,5)==3","Operation":"==","Value":"false"}],"branches":[{"branch":4,"result":"SendEmail(gc,'amp@gregzuro.com','greg@gregzuro.com','Big Brother is Watching', 'You have entered the building')"},{"branch":6,"result":"NoAction()"}]},{"branch":2,"decisions":[{"Inbranch":2,"Sequence":7,"trueBranch":5,"expression":"SlideingWindowGPSInArea(gc,5)==0 &&  SlideingWindowWifiInArea(gc,5)==0 && SlideingWindowIBeaconInArea(gc,5)==0","Operation":"==","Value":"true"},{"Inbranch":2,"Sequence":8,"trueBranch":7,"expression":"SlideingWindowGPSInArea(gc,5)==0 &&  SlideingWindowWifiInArea(gc,5)==0 && SlideingWindowIBeaconInArea(gc,5)==0","Operation":"==","Value":"false"}],"branches":[{"branch":5,"result":"SendEmail(gc,'amp@gregzuro.com','greg@gregzuro.com','Big Brother is Watching', 'Watch out for Obrien')"},{"branch":7,"result":"SendEmail(gc,'amp@gregzuro.com','greg@gregzuro.com','Big Brother is Watching', 'Watch out for Winston')"}]},{"branch":9,"decisions":[{"Inbranch":9,"Sequence":9,"trueBranch":11,"expression":"SlideingWindowGPSInArea(gc,5)==0 ||  SlideingWindowWifiInArea(gc,5)==0 || SlideingWindowIBeaconInArea(gc,5)==0","Operation":"==","Value":"true"},{"Inbranch":9,"Sequence":10,"trueBranch":10,"expression":"SlideingWindowGPSInArea(gc,5)==0 ||  SlideingWindowWifiInArea(gc,5)==0 || SlideingWindowIBeaconInArea(gc,5)==0","Operation":"==","Value":"false"}],"branches":[{"branch":11,"result":"ChangeState(gc,'ExitArea')"},{"branch":10,"result":"SendEmail(gc,'amp@gregzuro.com','greg@gregzuro.com','Big Brother is Watching', 'Watch out for Julia')"}]},{"branch":8,"decisions":[{"Inbranch":8,"Sequence":11,"trueBranch":12,"expression":"SlideingWindowGPSInArea(gc,5)==0 &&  SlideingWindowWifiInArea(gc,5)==0 && SlideingWindowIBeaconInArea(gc,5)==0","Operation":"==","Value":"true"},{"Inbranch":8,"Sequence":12,"trueBranch":13,"expression":"SlideingWindowGPSInArea(gc,5)==0 &&  SlideingWindowWifiInArea(gc,5)==0 && SlideingWindowIBeaconInArea(gc,5)==0","Operation":"==","Value":"false"}],"branches":[{"branch":12,"result":"ChangeState(gc,'OutOfArea')"},{"branch":13,"result":"ChangeState(gc,'InArea')"}]}]}]}
```

At this point `docker ps` should show running master, slave, and plugin containers. You can `tail -f` a file in `../var/log/` to monitor the activity in the cluster or observe plugin output directly with `docker attach sgms_plugin001`.

### Device Simulation

After the master, slave, and plugins have been set up, devices can be simulated to view data moving through the system. 

#### Local Device Switch

The easiest option for simulating devices is to run them locally in the same environment as master, slave, and plugin.  To acomplish this simply cd to the switch folder and run:
```
❯ make local-devices-up
```

The test device should be generating random location updates, which the plugin should be logging in large numbers.

#### Android

Clone the android project and build it:
```
git clone git@github.com:gregzuro/android-sdk-java.git
cd android-sdk-java
make
```

Go to the file sync/app/src/main/java/com/gregzuro/sync/sdk/service/DeviceSwitchImpl.java and ensure that ADDRESS_REMOTE_SLAVE and SLAVE_PORT are the same as your cluster.  If they are run the "app" module on your emulator or device.  Once permissions have been granted by the user, location data will be sent to the slave node on a regular interval.  You can view log data in Android's console.    


#### iOS

Clone the iOS project and build it:
```
git clone https://github.com/gregzuro/ios-sdk-objc.git
cd ios-sdk-objc/gregzuroSDK
make
```

Modify the device node startup parameters within SGNetworkingOperator.m file to match your cluster. Use XCode to run the contianer app on either the simulator or a test device.

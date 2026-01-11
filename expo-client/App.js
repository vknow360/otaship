import { StatusBar } from "expo-status-bar";
import { StyleSheet, Text, View, Image, Alert } from "react-native";
import Constants from "expo-constants";
import * as Updates from "expo-updates";
import { useEffect } from "react";

export default function App() {
  useEffect(() => {
    async function checkUpdates() {
      try {
        const update = await Updates.checkForUpdateAsync();
        console.log("Update available:", update.isAvailable);
        if (update.isAvailable) {
          await Updates.fetchUpdateAsync();
          Alert.alert("New Update Available", "Do you want to update now?", [
            {
              text: "Cancel",
              onPress: () => console.log("Cancel Pressed"),
              style: "cancel",
            },
            {
              text: "Update",
              onPress: () => Updates.reloadAsync(),
            },
          ]);
          // Don't auto-reload - let user decide when to restart
          console.log("Update downloaded. Will apply on next app restart.");
        }
      } catch (e) {
        console.log("Update check failed:", e);
      }
    }
    // Check for updates after app is fully loaded
    setTimeout(checkUpdates, 3000);
  }, []);
  return (
    <View style={styles.container}>
      <Text style={{ color: "black" }}>Change me...</Text>

      <Image source={require("./assets/favicon.png")} />
      <StatusBar style="auto" backgroundColor="black" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    alignItems: "center",
    justifyContent: "center",
  },
});

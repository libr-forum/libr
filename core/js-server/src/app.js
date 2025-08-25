import express from "express";
import { router } from "./routes/router.js";
const app = express();

app.use(express.json({ limit: "16kb" }));
app.use(express.urlencoded({ extended: true }));
app.use(express.static("public"));

app.get("/", (req, res) => {
    res.send("Hello World!");
});

app.use("/api", router);

export { app };
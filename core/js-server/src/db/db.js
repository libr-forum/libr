import mongoose from "mongoose";
import "dotenv/config";

const connectDB = async () => {
    try {
        console.log(process.env.MONGODB_URI)
        const connectionInstance = await mongoose.connect(process.env.MONGODB_URI);
        console.log(`\nMongoDB connected @ DB Host : ${connectionInstance.connection.host}`);
    } catch (error) {
        console.log("MongoDB connection failed", error);
        process.exit(1);
    }
};

export default connectDB;
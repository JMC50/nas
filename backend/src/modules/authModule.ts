import bcrypt from "bcryptjs";
import jwt from "jsonwebtoken";
import { 
    private_key, 
    AUTH_TYPE,
    PASSWORD_MIN_LENGTH,
    PASSWORD_REQUIRE_UPPERCASE,
    PASSWORD_REQUIRE_LOWERCASE,
    PASSWORD_REQUIRE_NUMBER,
    PASSWORD_REQUIRE_SPECIAL
} from "../config/config";
import getDatabase from "../sqlite";
import { intentList } from "../functions/auth";

export interface LocalAuthUser {
    userId: string;
    username: string;
    password?: string;
    krname?: string;
    global_name?: string;
    auth_type: 'local' | 'oauth';
}

export interface AuthResponse {
    success: boolean;
    token?: string;
    user?: any;
    message?: string;
}

export class AuthModule {
    /**
     * Validate password complexity
     */
    static validatePassword(password: string): { valid: boolean; message?: string } {
        if (password.length < PASSWORD_MIN_LENGTH) {
            return { valid: false, message: `Password must be at least ${PASSWORD_MIN_LENGTH} characters long` };
        }

        if (PASSWORD_REQUIRE_UPPERCASE && !/[A-Z]/.test(password)) {
            return { valid: false, message: "Password must contain at least one uppercase letter" };
        }

        if (PASSWORD_REQUIRE_LOWERCASE && !/[a-z]/.test(password)) {
            return { valid: false, message: "Password must contain at least one lowercase letter" };
        }

        if (PASSWORD_REQUIRE_NUMBER && !/[0-9]/.test(password)) {
            return { valid: false, message: "Password must contain at least one number" };
        }

        if (PASSWORD_REQUIRE_SPECIAL && !/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
            return { valid: false, message: "Password must contain at least one special character" };
        }

        return { valid: true };
    }

    /**
     * Hash password using bcrypt
     */
    static async hashPassword(password: string): Promise<string> {
        const salt = await bcrypt.genSalt(10);
        return bcrypt.hash(password, salt);
    }

    /**
     * Verify password against hash
     */
    static async verifyPassword(password: string, hash: string): Promise<boolean> {
        return bcrypt.compare(password, hash);
    }

    /**
     * Register a new user with ID/Password
     */
    static async registerLocalUser(
        userId: string, 
        username: string, 
        password: string, 
        krname?: string
    ): Promise<AuthResponse> {
        const db = getDatabase();

        // Check if auth type is allowed
        if (AUTH_TYPE === 'oauth') {
            return { 
                success: false, 
                message: "Local authentication is disabled. Please use OAuth." 
            };
        }

        // Check if user already exists
        const existingUser = db
            .prepare('SELECT * FROM users WHERE userId = ?')
            .get(userId);

        if (existingUser) {
            return { 
                success: false, 
                message: "User ID already exists" 
            };
        }

        // Validate password
        const passwordValidation = this.validatePassword(password);
        if (!passwordValidation.valid) {
            return { 
                success: false, 
                message: passwordValidation.message 
            };
        }

        // Hash password
        const hashedPassword = await this.hashPassword(password);

        // Insert new user
        try {
            const result = db.prepare(`
                INSERT INTO users (userId, username, password, krname, global_name, auth_type)
                VALUES (?, ?, ?, ?, ?, 'local')
            `).run(userId, username, hashedPassword, krname || '', username);

            // Create token
            const token = this.createToken(userId);

            return {
                success: true,
                token,
                user: {
                    userId,
                    username,
                    krname,
                    global_name: username,
                    auth_type: 'local'
                }
            };
        } catch (error) {
            return {
                success: false,
                message: "Failed to create user"
            };
        }
    }

    /**
     * Login with ID/Password
     */
    static async loginLocal(userId: string, password: string): Promise<AuthResponse> {
        const db = getDatabase();

        // Check if auth type is allowed
        if (AUTH_TYPE === 'oauth') {
            return { 
                success: false, 
                message: "Local authentication is disabled. Please use OAuth." 
            };
        }

        // Get user from database
        const user = db
            .prepare('SELECT * FROM users WHERE userId = ? AND auth_type = ?')
            .get(userId, 'local') as any;

        if (!user) {
            return { 
                success: false, 
                message: "Invalid user ID or password" 
            };
        }

        // Verify password
        const isValid = await this.verifyPassword(password, user.password);
        if (!isValid) {
            return { 
                success: false, 
                message: "Invalid user ID or password" 
            };
        }

        // Get user intents
        const intentsRow = db
            .prepare('SELECT intent FROM user_intents WHERE user_id = ?')
            .all(user.id) as any[];
        
        const intents = intentsRow.map(row => row.intent);

        // Create token
        const token = this.createToken(userId);

        return {
            success: true,
            token,
            user: {
                userId: user.userId,
                username: user.username,
                krname: user.krname,
                global_name: user.global_name || user.username,
                intents,
                auth_type: 'local'
            }
        };
    }

    /**
     * Change password for local user
     */
    static async changePassword(
        userId: string, 
        oldPassword: string, 
        newPassword: string
    ): Promise<AuthResponse> {
        const db = getDatabase();

        // Get user from database
        const user = db
            .prepare('SELECT * FROM users WHERE userId = ? AND auth_type = ?')
            .get(userId, 'local') as any;

        if (!user) {
            return { 
                success: false, 
                message: "User not found" 
            };
        }

        // Verify old password
        const isValid = await this.verifyPassword(oldPassword, user.password);
        if (!isValid) {
            return { 
                success: false, 
                message: "Invalid old password" 
            };
        }

        // Validate new password
        const passwordValidation = this.validatePassword(newPassword);
        if (!passwordValidation.valid) {
            return { 
                success: false, 
                message: passwordValidation.message 
            };
        }

        // Hash new password
        const hashedPassword = await this.hashPassword(newPassword);

        // Update password
        try {
            db.prepare('UPDATE users SET password = ? WHERE userId = ?')
                .run(hashedPassword, userId);

            return {
                success: true,
                message: "Password changed successfully"
            };
        } catch (error) {
            return {
                success: false,
                message: "Failed to change password"
            };
        }
    }

    /**
     * Create JWT token
     */
    static createToken(userId: string): string {
        const expires_in = Date.now() + 7 * 24 * 60 * 60 * 1000;
        return jwt.sign({
            userId,
            expires_in
        }, private_key);
    }

    /**
     * Get authentication configuration
     */
    static getAuthConfig() {
        return {
            authType: AUTH_TYPE,
            localAuthEnabled: AUTH_TYPE === 'local' || AUTH_TYPE === 'both',
            oauthEnabled: AUTH_TYPE === 'oauth' || AUTH_TYPE === 'both',
            passwordRequirements: {
                minLength: PASSWORD_MIN_LENGTH,
                requireUppercase: PASSWORD_REQUIRE_UPPERCASE,
                requireLowercase: PASSWORD_REQUIRE_LOWERCASE,
                requireNumber: PASSWORD_REQUIRE_NUMBER,
                requireSpecial: PASSWORD_REQUIRE_SPECIAL
            }
        };
    }
}
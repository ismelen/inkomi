import {
  GoogleSignin,
  isErrorWithCode,
  isSuccessResponse,
  statusCodes,
} from '@react-native-google-signin/google-signin';
import { googleWebClientId } from '../config';

let configured = false;

/** Google id token for POST /auth/google, or undefined when the user cancels. */
export async function getGoogleIdToken(): Promise<string | undefined> {
  if (!configured) {
    GoogleSignin.configure({ webClientId: googleWebClientId });
    configured = true;
  }

  try {
    await GoogleSignin.hasPlayServices({ showPlayServicesUpdateDialog: true });
    // Always show the account picker instead of reusing the previous account.
    await GoogleSignin.signOut();
    const response = await GoogleSignin.signIn();
    if (!isSuccessResponse(response)) return undefined;
    return response.data.idToken ?? undefined;
  } catch (err) {
    if (isErrorWithCode(err) && err.code === statusCodes.IN_PROGRESS) return undefined;
    throw err;
  }
}

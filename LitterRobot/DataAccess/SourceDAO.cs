using System;
using System.Collections.Generic;
using System.IdentityModel.Tokens.Jwt;
using System.Linq;
using System.Net.Http;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Primitives;
using LitterRobot.Models.Shared;
using Newtonsoft.Json;
using TwoMQTT.Core;
using TwoMQTT.Core.DataAccess;

namespace LitterRobot.DataAccess
{
    /// <summary>
    /// An class representing a managed way to interact with a source.
    /// </summary>
    public class SourceDAO : SourceDAO<SlugMapping, Command, Models.SourceManager.FetchResponse, object>
    {
        /// <summary>
        /// Initializes a new instance of the SourceDAO class.
        /// </summary>
        /// <param name="logger"></param>
        /// <param name="httpClientFactory"></param>
        /// <param name="cache"></param>
        /// <param name="apiKey"></param>
        /// <param name="login"></param>
        /// <param name="password"></param>
        /// <returns></returns>
        public SourceDAO(ILogger<SourceDAO> logger, IHttpClientFactory httpClientFactory, IMemoryCache cache, string login, string password) :
            base(logger)
        {
            this.Login = login;
            this.Password = password;
            this.Cache = cache;
            this.Client = httpClientFactory.CreateClient();
            this.ResponseObjCacheExpiration = new TimeSpan(0, 0, 17);
            this.LoginCacheExpiration = new TimeSpan(0, 51, 31);
        }

        /// <inheritdoc />
        public override async Task<Models.SourceManager.FetchResponse?> FetchOneAsync(SlugMapping key,
            CancellationToken cancellationToken = default)
        {
            try
            {
                return await this.FetchAsync(key.LRID, cancellationToken);
            }
            catch (Exception e)
            {
                var msg = e is HttpRequestException ? "Unable to fetch from the Litter Robot API" :
                          e is JsonException ? "Unable to deserialize response from the Litter Robot API" :
                          "Unable to send to the Litter Robot API";
                this.Logger.LogError(msg, e);
                return null;
            }
        }

        /// <inheritdoc />
        public override async Task<object?> SendOneAsync(Command item, CancellationToken cancellationToken = default)
        {
            try
            {
                var litterRobotId = item.Data.LitterRobotId;
                var command = this.TranslateCommand(item);
                return await this.SendAsync(litterRobotId, command, cancellationToken);
            }
            catch (Exception e)
            {
                var msg = e is HttpRequestException ? "Unable to send to the Litter Robot API" :
                          e is JsonException ? "Unable to serialize request to the Litter Robot API" :
                          "Unable to send to the Litter Robot API";
                this.Logger.LogError(msg, e);
                this.RemoveCachedLogin();
                return null;
            }
        }

        /// <summary>
        /// The internal cache.
        /// </summary>
        private readonly IMemoryCache Cache;

        /// <summary>
        /// The HTTP client used to access the source.
        /// </summary>
        private readonly HttpClient Client;

        /// <summary>
        /// The internal timeout for responses.
        /// </summary>
        private readonly TimeSpan ResponseObjCacheExpiration;

        /// <summary>
        /// The internal timeout for logins.
        /// </summary>
        private readonly TimeSpan LoginCacheExpiration;

        /// <summary>
        /// The Login to access the source.
        /// </summary>
        private readonly string Login;

        /// <summary>
        /// The Password to access the source.
        /// </summary>
        private readonly string Password;

        /// <summary>
        /// The semaphore to limit how many times Login is called.
        /// </summary>
        private readonly SemaphoreSlim LoginSemaphore = new SemaphoreSlim(1, 1);

        /// <summary>
        /// Get a request for the source
        /// </summary>
        /// <param name="method"></param>
        /// <param name="baseUrl"></param>
        /// <param name="obj"></param>
        /// <param name="token"></param>
        /// <returns></returns>
        private HttpRequestMessage Request(HttpMethod method, string baseUrl, object? obj, string token = "")
        {
            // Setup request + headers
            var request = new HttpRequestMessage(method, baseUrl);
            request.Headers.TryAddWithoutValidation("Content-Type", "application/json");
            if (!string.IsNullOrEmpty(token))
            {
                request.Headers.TryAddWithoutValidation("x-api-key", APIKEY);
                request.Headers.TryAddWithoutValidation("Authorization", token);
            }

            // Add optional content
            if (obj != null)
            {
                request.Content = new StringContent(JsonConvert.SerializeObject(obj), Encoding.UTF8, "application/json");
            }

            return request;
        }

        /// <summary>
        /// Login to the source
        /// </summary>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task<(string, string)> LoginAsync(CancellationToken cancellationToken = default)
        {
            this.Logger.LogDebug($"Started login proccess to LR");
            await this.LoginSemaphore.WaitAsync();

            try
            {
                // Try to get the login from cache
                if (this.Cache.TryGetValue(this.CacheKey(TYPEUSERID), out string userid) &&
                    this.Cache.TryGetValue(this.CacheKey(TYPETOKEN), out string token))
                {
                    this.Logger.LogDebug($"Found login credentials in cache");
                    return (userid, token);
                }

                // Hit the API
                var baseUrl = LOGINURL;
                var apiLogin = new List<KeyValuePair<string, string>>
                {
                    new KeyValuePair<string, string>("client_id", OAUTH_CLIENT_ID),
                    new KeyValuePair<string, string>("client_secret", OAUTH_CLIENT_SECRET),
                    new KeyValuePair<string, string>("grant_type", "password"),
                    new KeyValuePair<string, string>("username", this.Login),
                    new KeyValuePair<string, string>("password", this.Password),
                };
                //var request = this.Request(HttpMethod.Post, baseUrl, apiLogin);
                var resp = await this.Client.PostAsync(baseUrl, new FormUrlEncodedContent(apiLogin), cancellationToken);
                resp.EnsureSuccessStatusCode();
                var content = await resp.Content.ReadAsStringAsync();
                var obj = JsonConvert.DeserializeObject<APILoginResponse>(content);
                var jwt = new JwtSecurityToken(jwtEncodedString: obj.Access_Token);
                var jwtUserId = jwt.Claims.FirstOrDefault(x => x.Type == "userId").Value;
                var accessToken = obj.Access_Token;

                this.CacheLogin(jwtUserId, accessToken);

                this.Logger.LogDebug($"Finished login proccess to LR");
                return (jwtUserId, accessToken);
            }
            finally
            {
                this.LoginSemaphore.Release();
            }
        }

        /// <summary>
        /// Fetch one response from the source
        /// </summary>
        /// <param name="litterRobotId"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task<Models.SourceManager.FetchResponse?> FetchAsync(string litterRobotId,
            CancellationToken cancellationToken = default)
        {
            this.Logger.LogDebug($"Started finding {litterRobotId} from LR");
            // Check cache first to avoid hammering the Litter Robot API
            if (this.Cache.TryGetValue(this.CacheKey(TYPELRID, litterRobotId),
                out Models.SourceManager.FetchResponse cachedObj))
            {
                this.Logger.LogDebug($"Found {litterRobotId} from in the cache");
                return cachedObj;
            }

            var (userid, token) = await this.LoginAsync(cancellationToken);
            var baseUrl = string.Format(STATUSURL, userid);
            var request = this.Request(HttpMethod.Get, baseUrl, null, token);
            var resp = await this.Client.SendAsync(request, cancellationToken);
            resp.EnsureSuccessStatusCode();
            var content = await resp.Content.ReadAsStringAsync();
            var objs = JsonConvert.DeserializeObject<List<Models.SourceManager.FetchResponse>>(content);
            Models.SourceManager.FetchResponse? specificObj = null;
            foreach (var obj in objs)
            {
                // Cache all; return the specific one requested
                this.CacheResponse(obj);
                if (obj.LitterRobotId == litterRobotId)
                {
                    specificObj = obj;
                };
            }

            this.Logger.LogDebug($"Finished finding {litterRobotId} from LR");
            return specificObj;
        }

        /// <summary>
        /// Send one command to the source
        /// </summary>
        /// <param name="litterRobotId"></param>
        /// <param name="command"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task<object?> SendAsync(string litterRobotId, string command,
            CancellationToken cancellationToken = default)
        {
            this.Logger.LogDebug($"Started sending command {command} - {litterRobotId} to LR");

            var (userid, token) = await this.LoginAsync(cancellationToken);
            var baseUrl = string.Format(COMMANDURL, userid, litterRobotId);
            var apiCommand = new APICommand { command = command, litterRobotId = litterRobotId, };
            var request = this.Request(HttpMethod.Post, baseUrl, apiCommand, token);
            var resp = await this.Client.SendAsync(request, cancellationToken);
            resp.EnsureSuccessStatusCode();

            this.Logger.LogDebug($"Finished sending command {command} - {litterRobotId} to LR");
            return new object();
        }

        private readonly Dictionary<int, string> CommandMapping = new Dictionary<int, string>
        {
            { (int)CommandType.Power, "<P" },
            { (int)CommandType.Cycle, "<C" },
            { (int)CommandType.NightLight, "<N" },
            { (int)CommandType.PanelLock, "<L" },
            { (int)CommandType.WaitTime, "<W" },
            { (int)CommandType.Sleep, "<S" },
        };

        /// <summary>
        /// Translate interal commands w/data to something the Litter Robot API can utilize
        /// </summary>
        /// <param name="item"></param>
        /// <returns></returns>
        private string TranslateCommand(Command item)
        {
            this.Logger.LogDebug($"Started translating command {item}");

            // Covert from the internal Command into something LitterRobot knows about
            // Represent true = "1", false = "0"
            Func<bool, string> onoff = (bool x) => x ? "1" : "0";
            // Represent the number as hex; 7 is the default for bad numbers
            Func<long, string> waitTime = (long x) => x == 3 || x == 7 ? x.ToString() : x == 15 ? "F" : "7";
            // Represent off = "0", on = now, or any time (which is when the litter robot last slept)
            Func<string, string> sleepTime = (string x) => x == Const.OFF ? "0" : x == Const.ON ? "100:00:00" : $"1{x}";
            var cmd =
                (this.CommandMapping.ContainsKey(item.Command) ? this.CommandMapping[item.Command] : string.Empty) +
                (
                    item.Command == (int)CommandType.Power ? onoff(item.Data.Power) :
                    item.Command == (int)CommandType.Cycle ? onoff(item.Data.Cycle) :
                    item.Command == (int)CommandType.NightLight ? onoff(item.Data.NightLightActive) :
                    item.Command == (int)CommandType.PanelLock ? onoff(item.Data.PanelLockActive) :
                    item.Command == (int)CommandType.WaitTime ? waitTime(item.Data.CleanCycleWaitTimeMinutes) :
                    item.Command == (int)CommandType.Sleep ? sleepTime(item.Data.SleepMode) :
                    string.Empty
                );

            this.Logger.LogDebug($"Finished translating command {item} into {cmd}");
            return cmd;
        }

        /// <summary>
        /// Cache the login
        /// </summary>
        /// <param name="userid"></param>
        /// <param name="token"></param>
        private void CacheLogin(string userid, string token)
        {
            var cts = new CancellationTokenSource(this.LoginCacheExpiration);
            var cacheOpts = new MemoryCacheEntryOptions()
                 .AddExpirationToken(new CancellationChangeToken(cts.Token));

            this.Logger.LogDebug($"Caching LR Login");
            this.Cache.Set(this.CacheKey(TYPEUSERID), userid, cacheOpts);
            this.Cache.Set(this.CacheKey(TYPETOKEN), token, cacheOpts);
        }

        /// <summary>
        /// Cache the login
        /// </summary>
        private void RemoveCachedLogin()
        {
            this.Cache.Remove(this.CacheKey(TYPEUSERID));
            this.Cache.Remove(this.CacheKey(TYPETOKEN));
        }

        /// <summary>
        /// Cache the response
        /// </summary>
        /// <param name="obj"></param>
        private void CacheResponse(Models.SourceManager.FetchResponse obj)
        {
            var cts = new CancellationTokenSource(this.ResponseObjCacheExpiration);
            var cacheOpts = new MemoryCacheEntryOptions()
                 .AddExpirationToken(new CancellationChangeToken(cts.Token));

            this.Logger.LogDebug($"Started finding {obj.LitterRobotId} from LR");
            this.Cache.Set(this.CacheKey(TYPELRID, obj.LitterRobotId), obj, cacheOpts);
        }

        /// <summary>
        /// Generate a cache key
        /// </summary>
        /// <param name="type"></param>
        /// <param name="key"></param>
        /// <returns></returns>
        private string CacheKey(string type, string key = "KEY") => $"{type}_{key}";

        /// <summary>
        /// The base API url to access the source.
        /// </summary>
        private const string APIURL = "https://v2.api.whisker.iothings.site";

        /// <summary>
        /// The url to login to the source.
        /// </summary>
        private const string LOGINURL = "https://autopets.sso.iothings.site/oauth/token";

        /// <summary>
        /// The url to get status from the source.
        /// </summary>
        private const string STATUSURL = APIURL + "/users/{0}/robots";

        /// <summary>
        /// The url to send commands to access the source.
        /// </summary>
        private const string COMMANDURL = APIURL + "/users/{0}/robots/{1}/dispatch-commands";

        /// <summary>
        /// The key to cache litter robot objects.
        /// </summary>
        private const string TYPELRID = "LRID";

        /// <summary>
        /// The key to cache the userid.
        /// </summary>
        private const string TYPEUSERID = "USERID";

        /// <summary>
        /// The key to cache the token.
        /// </summary>
        private const string TYPETOKEN = "TOKEN";

        private const string APIKEY = "p7ndMoj61npRZP5CVz9v4Uj0bG769xy6758QRBPb";
        private const string OAUTH_CLIENT_ID = "IYXzWN908psOm7sNpe4G.ios.whisker.robots";
        private const string OAUTH_CLIENT_SECRET = "C63CLXOmwNaqLTB2xXo6QIWGwwBamcPuaul";

        /// <summary>
        /// 
        /// </summary>
        private class APILogin
        {
            public string client_id { get; set; } = string.Empty;
            public string client_secret { get; set; } = string.Empty;
            public string grant_type { get; set; } = string.Empty;
            public string username { get; set; } = string.Empty;
            public string password { get; set; } = string.Empty;
        }

        /// <summary>
        /// 
        /// </summary>
        private class APICommand
        {
            public string command { get; set; } = string.Empty;
            public string litterRobotId { get; set; } = string.Empty;
        }

        /// <summary>
        /// 
        /// </summary>
        private class APILoginResponse
        {

            public string Access_Token { get; set; } = string.Empty;
            public string Refresh_Token { get; set; } = string.Empty;
            public int Expires_In { get; set; } = 3600;
        }
    }
}

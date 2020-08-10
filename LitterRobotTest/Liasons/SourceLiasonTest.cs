using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Microsoft.VisualStudio.TestTools.UnitTesting;
using Moq;
using LitterRobot.DataAccess;
using LitterRobot.Liasons;
using LitterRobot.Models.Options;
using LitterRobot.Models.Shared;

namespace LitterRobotTest.Liasons
{
    [TestClass]
    public class SourceLiasonTest
    {
        [TestMethod]
        public async Task FetchAllAsyncTest()
        {
            var tests = new[] {
                new {
                    Q = new SlugMapping { LRID = BasicLRID, Slug = BasicSlug },
                    Expected = new { LRID = BasicLRID, UnitStatus = BasicUnitStatus }
                },
            };

            foreach (var test in tests)
            {
                var logger = new Mock<ILogger<SourceLiason>>();
                var sourceDAO = new Mock<ISourceDAO>();
                var opts = Options.Create(new SourceOpts());
                var sharedOpts = Options.Create(new SharedOpts
                {
                    Resources = new[] { test.Q }.ToList(),
                });

                sourceDAO.Setup(x => x.FetchOneAsync(test.Q, It.IsAny<CancellationToken>()))
                     .ReturnsAsync(new LitterRobot.Models.Source.Response
                     {
                         LitterRobotId = test.Expected.LRID,
                         SleepModeActive = "false",
                         UnitStatus = test.Expected.UnitStatus,
                     });

                var sourceLiason = new SourceLiason(logger.Object, sourceDAO.Object, opts, sharedOpts);
                await foreach (var result in sourceLiason.FetchAllAsync())
                {
                    Assert.AreEqual(test.Expected.LRID, result.LitterRobotId);
                    Assert.AreEqual(test.Expected.UnitStatus, result.UnitStatus);
                }
            }
        }

        private static string BasicSlug = "totallyaslug";
        private static string BasicUnitStatus = "Rdy";
        private static string BasicLRID = "15873525";
    }
}
